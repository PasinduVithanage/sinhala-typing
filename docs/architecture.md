# Sinhala Typing Assistant — Complete Architecture & Implementation Plan

> Windows-first Sinhala Unicode input engine built with Go + Wails + React

---

## Table of Contents

1. [Sinhala Unicode Deep Dive](#1-sinhala-unicode-deep-dive)
2. [Full Project Architecture](#2-full-project-architecture)
3. [Sinhala Input Engine](#3-sinhala-input-engine)
4. [Windows Keyboard Hooking System](#4-windows-keyboard-hooking-system)
5. [Unicode Normalization Engine](#5-unicode-normalization-engine)
6. [System-Wide Text Processing](#6-system-wide-text-processing)
7. [Desktop Application Design](#7-desktop-application-design)
8. [Performance Optimization](#8-performance-optimization)
9. [Security & Stability](#9-security--stability)
10. [MVP Development Roadmap](#10-mvp-development-roadmap)
11. [Mapping & Database Design](#11-mapping--database-design)
12. [UI/UX Design](#12-uiux-design)
13. [Future AI Features](#13-future-ai-features)
14. [Development Best Practices](#14-development-best-practices)

---

## 1. Sinhala Unicode Deep Dive

Before touching any code, you must deeply understand Sinhala Unicode or you will build a broken engine.

### 1.1 The Sinhala Unicode Block (U+0D80–U+0DFF)

```
CATEGORY              RANGE          EXAMPLES
─────────────────────────────────────────────────────────────────────
Independent Vowels    U+0D85–U+0D96  අ ආ ඇ ඈ ඉ ඊ උ ඌ ඍ ඎ ඏ ඐ එ ඒ ඓ ඔ ඕ ඖ
Consonants            U+0D9A–U+0DC6  ක ඛ ග ඝ ඞ ඟ ච ඡ ජ ඣ ඤ ඥ ට ඨ ඩ ඪ ණ ඬ
                                     ත ථ ද ධ න ඳ ප ඵ බ භ ම ඹ ය ර ල ව ශ ෂ ස හ ළ ෆ
Vowel Signs           U+0DCF–U+0DDF  ා ි ී ු ූ ෘ ෙ ේ ෛ ො ෝ ෞ
Special               U+0DCA         ් (Al-lakuna / Virama / Hal Kirima)
Digits                U+0DE6–U+0DEF  ෦ ෧ ෨ ෩ ෪ ෫ ෬ ෭ ෮ ෯
```

### 1.2 The Five Critical Rendering Problems

#### Problem 1: Kombuwa (ෙ U+0DD9) — Pre-base Vowel Sign

**The hardest problem.** Kombuwa visually appears to the LEFT of the consonant but Unicode mandates it is stored AFTER:

```
Visual:   ෙ + ක  →  ෙක
Unicode:  ක + ෙ  →  කෙ  ✓ CORRECT
          ෙ + ක  →  ෙක  ✗ WRONG (legacy typing order)
```

When a user types using a legacy layout and produces ෙ before ක, your engine must reorder it. Compound forms:

```
ො = ක + ෙ + ා  (kombuwa + aa sign)
ෝ = ක + ෙ + ා  (kombuwa + aa — long form)
ෞ = ක + ෛ      (au sign)
```

The Unicode standard defines these as sequences — no precomposed forms.

#### Problem 2: Ispili (ි U+0DD2) and Papili (ී U+0DD3) — The Gap Problem

The gap issue occurs when the consonant glyph and the i-matra glyph are not properly kerned:

```
ක + ි = කි   ← correct order, correct rendering
ක  ි  = ක ි  ← gap (matra stored before consonant, or renderer broken)
```

Root causes:
- Wrong Unicode order (matra stored before consonant)
- Font missing OpenType GSUB/GPOS rules for the consonant+matra pair
- Renderer not applying OTLG Sinhala shaping rules

Your engine's responsibility: ensure Unicode ORDER is correct (matra always after consonant).

#### Problem 3: Al-Lakuna Clusters (් U+0DCA)

Al-lakuna suppresses the inherent vowel of a consonant, enabling consonant clusters:

```
FORM 1 — Non-touching stack (default):
  ක + ් + ෂ = ක්ෂ  (ka + al-lakuna + sha)

FORM 2 — Touching/Conjunct form (ZWJ):
  ක + ් + ZWJ + ෂ = ක්‍ෂ  (ZWJ forces touching conjunct form)

FORM 3 — ZWNJ (explicit non-touching):
  ක + ් + ZWNJ + ෂ  (forces separate rendering)
```

#### Problem 4: Rakar (Subscript Ra)

Rakar is a subscript ra form that appears below/after a consonant:

```
Unicode: consonant + ් + ZWJ + ර
Example: ක + ් + ZWJ + ර = ක්‍ර  (kra)
         ත + ් + ZWJ + ර = ත්‍ර  (tra)
```

Many legacy systems omit the ZWJ, causing ක්ර (separate) instead of a proper subscript form.

#### Problem 5: Repaya (Superscript Ra)

Repaya is ra that appears ABOVE the following consonant:

```
Unicode: ර + ් + ZWJ + consonant
Example: ර + ් + ZWJ + ක = ර්‍ක  (rka with ra above)
```

### 1.3 Correct Sinhala Syllable Structure (BNF)

```
syllable     ::= (consonant_cluster)? vowel_nucleus vowel_sign?
               | independent_vowel

consonant_cluster ::= consonant (al_lakuna (ZWJ)? consonant)*
                    | ra al_lakuna ZWJ consonant   /* repaya at start */

vowel_nucleus ::= consonant | independent_vowel

vowel_sign   ::= kombuwa? matra?   /* kombuwa must precede other matras */
               | special_combination

kombuwa      ::= U+0DD9
matra        ::= U+0DCF | U+0DD2 | U+0DD3 | U+0DD4 | U+0DD6
               | U+0DDA | U+0DDC | U+0DDD | U+0DDE
```

---

## 2. Full Project Architecture

### 2.1 Folder Structure

```
sinhala-assistant/
├── cmd/
│   └── main.go                    # Wails entry point
│
├── internal/
│   ├── engine/
│   │   ├── transliteration/
│   │   │   ├── engine.go          # Core transliteration state machine
│   │   │   ├── statemachine.go    # FSM implementation
│   │   │   ├── token.go           # Token types and trie tokenizer
│   │   │   ├── phonetic.go        # Phonetic mapping handler
│   │   │   └── wijesekera.go      # Wijesekera layout handler
│   │   │
│   │   ├── normalization/
│   │   │   ├── normalizer.go      # Main normalization pipeline
│   │   │   ├── detector.go        # Malformed Sinhala detector
│   │   │   ├── reorder.go         # Character reordering rules
│   │   │   ├── repair.go          # Broken text repair
│   │   │   └── legacy.go          # Legacy encoding converter
│   │   │
│   │   └── composition/
│   │       ├── composer.go        # Syllable composer
│   │       ├── cluster.go         # Consonant cluster handler
│   │       ├── vowel.go           # Vowel sign handler
│   │       └── validator.go       # Unicode sequence validator
│   │
│   ├── hook/
│   │   ├── keyboard/
│   │   │   ├── hook.go            # WH_KEYBOARD_LL implementation
│   │   │   ├── filter.go          # Key event filter/router
│   │   │   ├── buffer.go          # Input buffer manager
│   │   │   └── injector.go        # SendInput text injection
│   │   │
│   │   └── clipboard/
│   │       ├── monitor.go         # Clipboard change monitor
│   │       ├── fixer.go           # Clipboard content fixer
│   │       └── history.go         # Clipboard history tracker
│   │
│   ├── window/
│   │   ├── detector.go            # Active window / focused field detection
│   │   ├── accessibility.go       # UI Automation wrapper
│   │   └── whitelist.go           # App compatibility whitelist
│   │
│   ├── tray/
│   │   ├── tray.go                # System tray implementation
│   │   ├── menu.go                # Tray menu builder
│   │   └── notifications.go       # Toast notifications
│   │
│   ├── config/
│   │   ├── config.go              # Configuration manager
│   │   ├── schema.go              # Config schema/validation
│   │   └── hotkeys.go             # Hotkey configuration
│   │
│   ├── ipc/
│   │   ├── bridge.go              # Wails IPC bridge
│   │   ├── events.go              # Event definitions
│   │   └── handlers.go            # IPC method handlers
│   │
│   └── logging/
│       ├── logger.go              # Structured logger
│       └── debug.go               # Debug/diagnostic tools
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── TrayPanel/         # Main tray popup panel
│   │   │   ├── Settings/          # Settings screens
│   │   │   ├── KeyboardLayout/    # Visual keyboard layout
│   │   │   ├── SuggestionPopup/   # Typing suggestions overlay
│   │   │   └── StatusBar/         # Engine status indicator
│   │   │
│   │   ├── hooks/
│   │   │   ├── useEngine.ts       # Engine state hook
│   │   │   ├── useHotkeys.ts      # Hotkey listener hook
│   │   │   └── useClipboard.ts    # Clipboard monitor hook
│   │   │
│   │   ├── store/
│   │   │   ├── engineStore.ts     # Zustand engine state
│   │   │   ├── settingsStore.ts   # Settings state
│   │   │   └── historyStore.ts    # Typing history state
│   │   │
│   │   ├── wailsjs/               # Auto-generated Wails bindings
│   │   ├── App.tsx
│   │   └── main.tsx
│   │
│   ├── tailwind.config.js
│   └── package.json
│
├── mappings/
│   ├── phonetic.json              # Phonetic transliteration rules
│   ├── wijesekera.json            # Wijesekera layout map
│   ├── sinhala_unicode.json       # Unicode character database
│   └── normalization_rules.json   # Normalization correction rules
│
├── assets/
│   ├── icons/                     # Tray icons (active / inactive / error)
│   └── fonts/                     # Bundled Sinhala fonts
│
├── wails.json
├── go.mod
└── go.sum
```

### 2.2 Module Dependency Graph

```
cmd/main.go
    │
    ├── internal/ipc/bridge.go  ←→  frontend (React)
    │       │
    │       ├── internal/engine/transliteration/  (pure Go, no syscalls)
    │       ├── internal/engine/normalization/     (pure Go, no syscalls)
    │       ├── internal/engine/composition/       (pure Go, no syscalls)
    │       │
    │       ├── internal/hook/keyboard/   (Windows API)
    │       ├── internal/hook/clipboard/  (Windows API)
    │       ├── internal/window/          (Windows API)
    │       │
    │       ├── internal/tray/            (Windows API)
    │       └── internal/config/          (filesystem)
    │
    └── internal/logging/
```

> **Critical design rule**: The `engine/` packages are pure Go with zero Windows API calls. This enables unit testing without Windows and future cross-platform support.

---

## 3. Sinhala Input Engine

### 3.1 Transliteration State Machine

```go
// internal/engine/transliteration/statemachine.go

package transliteration

type State int

const (
    StateIdle          State = iota
    StateConsonant            // typed a consonant, awaiting vowel or next consonant
    StateConsonantCluster     // typed hal trigger, awaiting next consonant
    StateVowelSign            // applied a vowel sign
    StateComplete             // syllable complete, ready to commit
)

type Event int

const (
    EventConsonant   Event = iota
    EventVowel             // independent vowel
    EventVowelSign         // vowel sign/matra
    EventHal               // al-lakuna trigger
    EventKombuwa           // ෙ specifically (needs pre-base handling)
    EventSpace
    EventBackspace
    EventOther
)

type Token struct {
    Type    Event
    Sinhala rune   // Sinhala Unicode codepoint
    Raw     string // original keystrokes that produced this token
}

type TransliterationResult struct {
    Output      string // Sinhala Unicode string to emit
    DeleteCount int    // how many raw chars to delete before inserting Output
    Committed   bool   // whether this is a final commit or tentative
}

type StateMachine struct {
    state   State
    buffer  []Token
    mapping MappingTable
    config  Config
}

func NewStateMachine(mapping MappingTable, config Config) *StateMachine {
    return &StateMachine{
        state:   StateIdle,
        buffer:  make([]Token, 0, 8),
        mapping: mapping,
        config:  config,
    }
}

func (sm *StateMachine) ProcessKey(rawInput string) (TransliterationResult, bool) {
    token, ok := sm.mapping.Lookup(rawInput)
    if !ok {
        result := sm.commit()
        sm.reset()
        return result, true
    }
    sm.buffer = append(sm.buffer, token)
    return sm.transition(token)
}

func (sm *StateMachine) transition(t Token) (TransliterationResult, bool) {
    switch sm.state {
    case StateIdle:
        return sm.fromIdle(t)
    case StateConsonant:
        return sm.fromConsonant(t)
    case StateConsonantCluster:
        return sm.fromCluster(t)
    case StateVowelSign:
        return sm.fromVowelSign(t)
    }
    return TransliterationResult{}, false
}

func (sm *StateMachine) fromIdle(t Token) (TransliterationResult, bool) {
    switch t.Type {
    case EventConsonant:
        sm.state = StateConsonant
        return TransliterationResult{
            Output:      string(t.Sinhala),
            DeleteCount: len(t.Raw),
            Committed:   false, // tentative — vowel sign may follow
        }, true

    case EventVowel:
        sm.state = StateComplete
        return TransliterationResult{
            Output:      string(t.Sinhala),
            DeleteCount: len(t.Raw),
            Committed:   true,
        }, true
    }
    return TransliterationResult{}, false
}

func (sm *StateMachine) fromConsonant(t Token) (TransliterationResult, bool) {
    switch t.Type {
    case EventVowelSign:
        sm.state = StateVowelSign
        composed := sm.composeVowelSign(sm.lastConsonant(), t.Sinhala)
        return TransliterationResult{
            Output:      composed,
            DeleteCount: sm.rawBufferLength(),
            Committed:   true,
        }, true

    case EventKombuwa:
        sm.state = StateVowelSign
        // Store kombuwa — more vowel signs may follow (ො = kombuwa + ා)
        return TransliterationResult{
            Output:      string(sm.lastConsonant()) + string(t.Sinhala),
            DeleteCount: sm.rawBufferLength(),
            Committed:   false,
        }, true

    case EventHal:
        sm.state = StateConsonantCluster
        return TransliterationResult{}, false // wait for next consonant

    case EventConsonant:
        // Two consonants in a row — commit first, start new syllable
        prev := sm.commitBuffer()
        sm.reset()
        sm.buffer = append(sm.buffer, t)
        sm.state = StateConsonant
        return TransliterationResult{
            Output:      prev + string(t.Sinhala),
            DeleteCount: sm.rawBufferLength(),
            Committed:   false,
        }, true
    }
    return TransliterationResult{}, false
}

func (sm *StateMachine) fromCluster(t Token) (TransliterationResult, bool) {
    if t.Type != EventConsonant {
        result := sm.commit()
        sm.reset()
        return result, true
    }
    cluster := sm.buildCluster(t.Sinhala)
    sm.state = StateConsonant
    return TransliterationResult{
        Output:      cluster,
        DeleteCount: sm.rawBufferLength(),
        Committed:   false,
    }, true
}

const (
    AlLakuna = '්' // ්
    ZWJ      = '‍' // Zero Width Joiner
    ZWNJ     = '‌' // Zero Width Non-Joiner
    Ra       = 'ර' // ර
)

func (sm *StateMachine) buildCluster(next rune) string {
    prev := sm.lastConsonant()
    if prev == Ra && sm.config.UseZWJClusters {
        // Repaya form: ර + ් + ZWJ + consonant
        return string(prev) + string(AlLakuna) + string(ZWJ) + string(next)
    }
    if next == Ra && sm.config.UseZWJClusters {
        // Rakar form: consonant + ් + ZWJ + ර
        return string(prev) + string(AlLakuna) + string(ZWJ) + string(next)
    }
    return string(prev) + string(AlLakuna) + string(next)
}

func (sm *StateMachine) composeVowelSign(consonant, sign rune) string {
    if sm.hasKombuwa() {
        return string(consonant) + string('ෙ') + string(sign)
    }
    return string(consonant) + string(sign)
}

func (sm *StateMachine) lastConsonant() rune {
    for i := len(sm.buffer) - 1; i >= 0; i-- {
        if sm.buffer[i].Type == EventConsonant {
            return sm.buffer[i].Sinhala
        }
    }
    return 0
}

func (sm *StateMachine) hasKombuwa() bool {
    for _, t := range sm.buffer {
        if t.Sinhala == 'ෙ' {
            return true
        }
    }
    return false
}

func (sm *StateMachine) rawBufferLength() int {
    total := 0
    for _, t := range sm.buffer {
        total += len(t.Raw)
    }
    return total
}

func (sm *StateMachine) commitBuffer() string {
    var out []rune
    for _, t := range sm.buffer {
        out = append(out, t.Sinhala)
    }
    return string(out)
}

func (sm *StateMachine) commit() TransliterationResult {
    return TransliterationResult{
        Output:      sm.commitBuffer(),
        DeleteCount: sm.rawBufferLength(),
        Committed:   true,
    }
}

func (sm *StateMachine) reset() {
    sm.state = StateIdle
    sm.buffer = sm.buffer[:0]
}
```

### 3.2 Phonetic Mapping Table

```go
// internal/engine/transliteration/phonetic.go

package transliteration

type MappingEntry struct {
    Output rune
    Type   Event
    MaxLen int
}

// Core phonetic mapping — English keystrokes → Sinhala Unicode
var PhoneticMap = map[string]MappingEntry{
    // Independent vowels
    "a":  {0x0D85, EventVowel, 1},   // අ
    "aa": {0x0D86, EventVowel, 2},   // ආ
    "A":  {0x0D86, EventVowel, 1},   // ආ
    "ae": {0x0D87, EventVowel, 2},   // ඇ
    "i":  {0x0D89, EventVowel, 1},   // ඉ
    "ii": {0x0D8A, EventVowel, 2},   // ඊ
    "u":  {0x0D8B, EventVowel, 1},   // උ
    "uu": {0x0D8C, EventVowel, 2},   // ඌ
    "e":  {0x0D91, EventVowel, 1},   // එ
    "ee": {0x0D92, EventVowel, 2},   // ඒ
    "o":  {0x0D94, EventVowel, 1},   // ඔ
    "oo": {0x0D95, EventVowel, 2},   // ඕ

    // Consonants
    "k":  {0x0D9A, EventConsonant, 1},  // ක
    "kh": {0x0D9B, EventConsonant, 2},  // ඛ
    "g":  {0x0D9C, EventConsonant, 1},  // ග
    "gh": {0x0D9D, EventConsonant, 2},  // ඝ
    "ng": {0x0D9E, EventConsonant, 2},  // ඞ
    "c":  {0x0DA0, EventConsonant, 1},  // ච
    "ch": {0x0DA0, EventConsonant, 2},  // ච
    "j":  {0x0DA2, EventConsonant, 1},  // ජ
    "jh": {0x0DA3, EventConsonant, 2},  // ඣ
    "ny": {0x0DA4, EventConsonant, 2},  // ඤ
    "T":  {0x0DA7, EventConsonant, 1},  // ට (retroflex)
    "Th": {0x0DA8, EventConsonant, 2},  // ඨ
    "D":  {0x0DA9, EventConsonant, 1},  // ඩ
    "Dh": {0x0DAA, EventConsonant, 2},  // ඪ
    "N":  {0x0DAB, EventConsonant, 1},  // ණ
    "t":  {0x0DAD, EventConsonant, 1},  // ත
    "th": {0x0DAE, EventConsonant, 2},  // ථ
    "d":  {0x0DAF, EventConsonant, 1},  // ද
    "dh": {0x0DB0, EventConsonant, 2},  // ධ
    "n":  {0x0DB1, EventConsonant, 1},  // න
    "p":  {0x0DB4, EventConsonant, 1},  // ප
    "ph": {0x0DB5, EventConsonant, 2},  // ඵ
    "b":  {0x0DB6, EventConsonant, 1},  // බ
    "bh": {0x0DB7, EventConsonant, 2},  // භ
    "m":  {0x0DB8, EventConsonant, 1},  // ම
    "mb": {0x0DB9, EventConsonant, 2},  // ඹ
    "y":  {0x0DBA, EventConsonant, 1},  // ය
    "r":  {0x0DBB, EventConsonant, 1},  // ර
    "l":  {0x0DBD, EventConsonant, 1},  // ල
    "v":  {0x0DC0, EventConsonant, 1},  // ව
    "w":  {0x0DC0, EventConsonant, 1},  // ව
    "sh": {0x0DC1, EventConsonant, 2},  // ශ
    "Sh": {0x0DC2, EventConsonant, 2},  // ෂ
    "s":  {0x0DC3, EventConsonant, 1},  // ස
    "h":  {0x0DC4, EventConsonant, 1},  // හ
    "L":  {0x0DC5, EventConsonant, 1},  // ළ
    "f":  {0x0DC6, EventConsonant, 1},  // ෆ

    // Kombuwa
    "E":  {0x0DD9, EventKombuwa, 1},   // ෙ

    // Al-lakuna triggers
    "H":  {AlLakuna, EventHal, 1},     // ්
    "^":  {AlLakuna, EventHal, 1},     // ්
}

// VowelSignMap: when in StateConsonant, these vowel keys become vowel signs
var VowelSignMap = map[rune]rune{
    'a': 0x0DCF, // ා
    'i': 0x0DD2, // ි (ispili)
    'I': 0x0DD3, // ී (papili)
    'u': 0x0DD4, // ු
    'U': 0x0DD6, // ූ
    'e': 0x0DD9, // ෙ (kombuwa)
    'E': 0x0DDA, // ේ
    'o': 0x0DDC, // ො (kombuwa + aa — handled as sequence)
}
```

### 3.3 Trie-Based Prefix Lookup

```go
// internal/engine/transliteration/token.go

package transliteration

type TrieNode struct {
    children map[rune]*TrieNode
    entry    *MappingEntry
}

type Trie struct {
    root *TrieNode
}

func NewTrie() *Trie {
    return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(key string, entry MappingEntry) {
    node := t.root
    for _, ch := range key {
        child, ok := node.children[ch]
        if !ok {
            child = &TrieNode{children: make(map[rune]*TrieNode)}
            node.children[ch] = child
        }
        node = child
    }
    node.entry = &entry
}

// LookupWithBuffer finds the longest matching prefix in the buffer.
// Returns (entry, matchLength, isPrefix) where isPrefix=true means
// the buffer could be the start of a longer valid key — don't commit yet.
func (t *Trie) LookupWithBuffer(buf string) (entry *MappingEntry, matchLen int, isPrefix bool) {
    node := t.root
    lastMatch := (*MappingEntry)(nil)
    lastMatchLen := 0

    for i, ch := range buf {
        child, ok := node.children[ch]
        if !ok {
            return lastMatch, lastMatchLen, false
        }
        node = child
        if node.entry != nil {
            lastMatch = node.entry
            lastMatchLen = i + 1
        }
    }

    isPrefix = len(node.children) > 0
    return lastMatch, lastMatchLen, isPrefix
}
```

---

## 4. Windows Keyboard Hooking System

### 4.1 Architecture Overview

```
User Keyboard Input
       │
       ▼
 Windows Kernel
       │
       ▼
 WH_KEYBOARD_LL Hook (your process)
       │
    ┌──┴──────────────────────────┐
    │                             │
    ▼                             ▼
 INTERCEPT                    PASS THROUGH
 (Sinhala mode ON)           (mode OFF or non-target app)
    │
    ▼
 Key Event Queue (goroutine-safe buffered channel)
    │
    ▼
 Input Buffer (accumulates keystrokes)
    │
    ▼
 Transliteration Engine
    │
    ▼
 Result: (Sinhala string, delete_count)
    │
    ▼
 Injector:
   1. Send Backspace × delete_count  (erase typed keys)
   2. Send Unicode chars via SendInput
```

### 4.2 Low-Level Keyboard Hook

```go
// internal/hook/keyboard/hook.go

package keyboard

import (
    "sync"
    "sync/atomic"
    "unsafe"

    "golang.org/x/sys/windows"
)

var (
    user32   = windows.NewLazySystemDLL("user32.dll")
    kernel32 = windows.NewLazySystemDLL("kernel32.dll")

    procSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
    procCallNextHookEx           = user32.NewProc("CallNextHookEx")
    procUnhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
    procGetMessageW              = user32.NewProc("GetMessageW")
    procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
    procSendInput                = user32.NewProc("SendInput")
    procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
)

const (
    WH_KEYBOARD_LL  = 13
    WM_KEYDOWN      = 0x0100
    WM_KEYUP        = 0x0101
    WM_SYSKEYDOWN   = 0x0104
    HC_ACTION       = 0
    LLKHF_INJECTED  = 0x10 // flag on events we inject — prevents infinite loop

    VK_BACK  = 0x08
    VK_SHIFT = 0x10
    VK_CTRL  = 0x11
    VK_MENU  = 0x12
)

type KBDLLHOOKSTRUCT struct {
    VkCode      uint32
    ScanCode    uint32
    Flags       uint32
    Time        uint32
    DwExtraInfo uintptr
}

type KeyEvent struct {
    VkCode    uint32
    ScanCode  uint32
    Flags     uint32
    Shift     bool
    Ctrl      bool
    Alt       bool
    IsKeyDown bool
}

type Hook struct {
    handle   windows.Handle
    events   chan KeyEvent
    active   atomic.Bool
    mu       sync.Mutex
    hookProc uintptr
    stopCh   chan struct{}
}

func New() *Hook {
    return &Hook{
        events: make(chan KeyEvent, 256),
        stopCh: make(chan struct{}),
    }
}

var globalHook *Hook

func (h *Hook) Start() error {
    h.mu.Lock()
    defer h.mu.Unlock()

    globalHook = h
    cb := windows.NewCallback(lowLevelKeyboardProc)
    h.hookProc = cb

    mod, _, _ := procGetModuleHandleW.Call(0)
    handle, _, err := procSetWindowsHookExW.Call(
        WH_KEYBOARD_LL, cb, mod, 0,
    )
    if handle == 0 {
        return err
    }
    h.handle = windows.Handle(handle)
    h.active.Store(true)
    return nil
}

func (h *Hook) Stop() {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.handle != 0 {
        procUnhookWindowsHookEx.Call(uintptr(h.handle))
        h.handle = 0
        h.active.Store(false)
    }
}

func (h *Hook) IsActive() bool { return h.active.Load() }
func (h *Hook) SetActive(v bool) { h.active.Store(v) }
func (h *Hook) Events() <-chan KeyEvent { return h.events }

func lowLevelKeyboardProc(nCode int, wParam, lParam uintptr) uintptr {
    if nCode != HC_ACTION || globalHook == nil {
        return callNext(nCode, wParam, lParam)
    }

    kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

    // CRITICAL: skip injected events to prevent infinite loop
    if kbd.Flags&LLKHF_INJECTED != 0 {
        return callNext(nCode, wParam, lParam)
    }

    if !globalHook.active.Load() {
        return callNext(nCode, wParam, lParam)
    }

    isDown := wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN

    event := KeyEvent{
        VkCode:    kbd.VkCode,
        ScanCode:  kbd.ScanCode,
        Flags:     kbd.Flags,
        IsKeyDown: isDown,
        Shift:     isKeyPressed(VK_SHIFT),
        Ctrl:      isKeyPressed(VK_CTRL),
        Alt:       isKeyPressed(VK_MENU),
    }

    select {
    case globalHook.events <- event:
        if isDown {
            return 1 // swallow the key
        }
    default:
        // Buffer full — pass through
    }

    return callNext(nCode, wParam, lParam)
}

func callNext(nCode int, wParam, lParam uintptr) uintptr {
    ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
    return ret
}

func isKeyPressed(vk uint32) bool {
    ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
    return ret&0x8000 != 0
}

func (h *Hook) RunMessageLoop() {
    type MSG struct {
        Hwnd    uintptr
        Message uint32
        WParam  uintptr
        LParam  uintptr
        Time    uint32
        Pt      [2]int32
    }
    var msg MSG
    for {
        select {
        case <-h.stopCh:
            return
        default:
        }
        ret, _, _ := procGetMessageW.Call(
            uintptr(unsafe.Pointer(&msg)), 0, 0, 0,
        )
        if ret == 0 || ret == ^uintptr(0) {
            return
        }
    }
}
```

### 4.3 Text Injection via SendInput

```go
// internal/hook/keyboard/injector.go

package keyboard

import (
    "unicode/utf16"
    "unsafe"
)

const (
    INPUT_KEYBOARD    = 1
    KEYEVENTF_UNICODE = 0x0004
    KEYEVENTF_KEYUP   = 0x0002
)

type KEYBDINPUT struct {
    WVk         uint16
    WScan       uint16
    DwFlags     uint32
    Time        uint32
    DwExtraInfo uintptr
}

type INPUT struct {
    Type uint32
    _    [4]byte
    Ki   KEYBDINPUT
    _    [8]byte
}

// InjectUnicode sends a Unicode string as keyboard input events.
// Each UTF-16 code unit becomes a WM_CHAR event.
// This is the most compatible method across Windows applications.
func InjectUnicode(text string) error {
    runes := []rune(text)
    u16 := utf16.Encode(runes)

    inputs := make([]INPUT, 0, len(u16)*2)
    for _, u := range u16 {
        inputs = append(inputs,
            INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WScan: u, DwFlags: KEYEVENTF_UNICODE}},
            INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WScan: u, DwFlags: KEYEVENTF_UNICODE | KEYEVENTF_KEYUP}},
        )
    }

    if len(inputs) == 0 {
        return nil
    }

    ret, _, err := procSendInput.Call(
        uintptr(len(inputs)),
        uintptr(unsafe.Pointer(&inputs[0])),
        unsafe.Sizeof(inputs[0]),
    )
    if ret == 0 {
        return err
    }
    return nil
}

// InjectBackspaces sends N backspace key events to erase previous input
func InjectBackspaces(n int) error {
    if n <= 0 {
        return nil
    }
    inputs := make([]INPUT, 0, n*2)
    for i := 0; i < n; i++ {
        inputs = append(inputs,
            INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_BACK}},
            INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_BACK, DwFlags: KEYEVENTF_KEYUP}},
        )
    }
    ret, _, err := procSendInput.Call(
        uintptr(len(inputs)),
        uintptr(unsafe.Pointer(&inputs[0])),
        unsafe.Sizeof(inputs[0]),
    )
    if ret == 0 {
        return err
    }
    return nil
}

// ReplaceWithSinhala is the main replacement operation:
// erase `deleteCount` characters, then inject the Sinhala unicode string.
func ReplaceWithSinhala(sinhalaText string, deleteCount int) error {
    if err := InjectBackspaces(deleteCount); err != nil {
        return err
    }
    return InjectUnicode(sinhalaText)
}
```

---

## 5. Unicode Normalization Engine

### 5.1 Malformed Sinhala Detector

```go
// internal/engine/normalization/detector.go

package normalization

import "golang.org/x/text/unicode/norm"

type Issue int

const (
    IssueNone              Issue = iota
    IssueKombuwaBeforeBase       // ෙ placed before consonant
    IssueVowelSignBeforeBase     // matra before consonant
    IssueDuplicateVirama         // ්් double virama
    IssueOrphanedVirama          // virama without preceding consonant
    IssueWrongClusterOrder       // cluster components misordered
    IssueLegacyEncoding          // detected legacy Sinhala font encoding
)

type Detection struct {
    StartIndex int
    EndIndex   int
    Issue      Issue
    Context    string
}

// Analyze scans a Sinhala string for rendering issues.
func Analyze(s string) []Detection {
    runes := []rune(norm.NFC.String(s))
    var issues []Detection

    for i := 0; i < len(runes); i++ {
        r := runes[i]

        // Kombuwa must follow a consonant
        if r == 0x0DD9 {
            if i == 0 || !isSinhalaConsonant(runes[i-1]) {
                issues = append(issues, Detection{
                    StartIndex: i, EndIndex: i + 1,
                    Issue:   IssueKombuwaBeforeBase,
                    Context: string(runes[max(0, i-1):min(len(runes), i+2)]),
                })
            }
        }

        // Any vowel sign must follow a consonant
        if isSinhalaVowelSign(r) && r != 0x0DD9 {
            if i == 0 || (!isSinhalaConsonant(runes[i-1]) && runes[i-1] != 0x0DCA) {
                issues = append(issues, Detection{
                    StartIndex: i, EndIndex: i + 1,
                    Issue:   IssueVowelSignBeforeBase,
                    Context: string(runes[max(0, i-1):min(len(runes), i+2)]),
                })
            }
        }

        // Double virama
        if r == 0x0DCA && i+1 < len(runes) && runes[i+1] == 0x0DCA {
            issues = append(issues, Detection{StartIndex: i, EndIndex: i + 2, Issue: IssueDuplicateVirama})
        }

        // Orphaned virama
        if r == 0x0DCA && (i == 0 || !isSinhalaConsonant(runes[i-1])) {
            issues = append(issues, Detection{StartIndex: i, EndIndex: i + 1, Issue: IssueOrphanedVirama})
        }
    }

    return issues
}

func isSinhalaConsonant(r rune) bool { return r >= 0x0D9A && r <= 0x0DC6 }
func isSinhalaVowelSign(r rune) bool  { return r >= 0x0DCF && r <= 0x0DDF }

func max(a, b int) int {
    if a > b { return a }
    return b
}
func min(a, b int) int {
    if a < b { return a }
    return b
}
```

### 5.2 Sinhala Text Repair Engine

```go
// internal/engine/normalization/repair.go

package normalization

import "golang.org/x/text/unicode/norm"

// Repair fixes known Sinhala Unicode ordering problems.
func Repair(s string) string {
    s = norm.NFC.String(s)
    runes := []rune(s)
    var out []rune

    i := 0
    for i < len(runes) {
        r := runes[i]

        // Fix: Kombuwa before consonant → swap
        if r == 0x0DD9 && i+1 < len(runes) && isSinhalaConsonant(runes[i+1]) {
            out = append(out, runes[i+1], r)
            i += 2
            continue
        }

        // Fix: Vowel sign before consonant → swap
        if isSinhalaVowelSign(r) && i+1 < len(runes) && isSinhalaConsonant(runes[i+1]) {
            out = append(out, runes[i+1], r)
            i += 2
            continue
        }

        // Fix: Double virama → single
        if r == 0x0DCA && i+1 < len(runes) && runes[i+1] == 0x0DCA {
            out = append(out, r)
            i += 2
            continue
        }

        // Fix: Orphaned virama → remove
        if r == 0x0DCA && (len(out) == 0 || !isSinhalaConsonant(out[len(out)-1])) {
            i++
            continue
        }

        // Fix: Rakar without ZWJ → add ZWJ
        // Pattern: consonant + ් + ර  →  consonant + ් + ZWJ + ර
        if r == 0x0DCA && i+1 < len(runes) && runes[i+1] == 0x0DBB {
            if i+2 >= len(runes) || runes[i+2] != 0x200D {
                out = append(out, r, 0x200D)
                i++
                continue
            }
        }

        out = append(out, r)
        i++
    }

    return norm.NFC.String(string(out))
}

// RepairString is the main entry point: detects legacy encoding first,
// then applies Unicode repair rules.
func RepairString(s string) string {
    if IsLikelyLegacyEncoded(s) {
        s = ConvertLegacyWijesekera(s)
    }
    if len(Analyze(s)) == 0 {
        return s
    }
    return Repair(s)
}
```

### 5.3 Wijesekera Legacy Converter

```go
// internal/engine/normalization/legacy.go

package normalization

import "strings"

// WijeskeeraToUnicode maps Wijesekera font ASCII positions to Sinhala Unicode
var WijeskeeraToUnicode = map[rune]rune{
    'q': 0x0DCA, // ්
    'w': 0x0DD9, // ෙ
    'e': 0x0DD0, // ැ
    'r': 0x0DBB, // ර
    't': 0x0DAD, // ත
    'y': 0x0DBA, // ය
    'u': 0x0DD4, // ු
    'i': 0x0DD2, // ි (ispili)
    'o': 0x0DDC, // ො
    'p': 0x0DB4, // ප
    'a': 0x0D85, // අ
    's': 0x0DC3, // ස
    'd': 0x0DAF, // ද
    'f': 0x0DC6, // ෆ
    'g': 0x0D9C, // ග
    'h': 0x0DC4, // හ
    'j': 0x0DA2, // ජ
    'k': 0x0D9A, // ක
    'l': 0x0DBD, // ල
    'z': 0x0DC6, // ෆ
    'x': 0x0D82, // ං
    'c': 0x0DA0, // ච
    'v': 0x0DC0, // ව
    'b': 0x0DB6, // බ
    'n': 0x0DB1, // න
    'm': 0x0DB8, // ම
    // Uppercase
    'Q': 0x0DDF, // ෟ
    'W': 0x0DDB, // ෛ
    'E': 0x0DD1, // ැ short
    'R': 0x0DBB, // ර
    'T': 0x0DA7, // ට
    'Y': 0x0DDE, // ෞ
    'U': 0x0DD6, // ූ
    'I': 0x0DD3, // ී (papili)
    'O': 0x0DDD, // ෝ
    'P': 0x0DB5, // ඵ
    'A': 0x0D86, // ආ
    'S': 0x0DC2, // ෂ
    'D': 0x0DA9, // ඩ
    'G': 0x0D9D, // ඝ
    'J': 0x0DA3, // ඣ
    'K': 0x0D9B, // ඛ
    'L': 0x0DC5, // ළ
    'N': 0x0DAB, // ණ
    'M': 0x0DB9, // ඹ
    ';': 0x0D83, // ඃ
    '\'': 0x0DCF, // ා
}

func ConvertLegacyWijesekera(s string) string {
    var b strings.Builder
    b.Grow(len(s) * 3) // Sinhala chars are 3 bytes in UTF-8
    for _, r := range s {
        if mapped, ok := WijeskeeraToUnicode[r]; ok {
            b.WriteRune(mapped)
        } else {
            b.WriteRune(r)
        }
    }
    return b.String()
}

func IsLikelyLegacyEncoded(s string) bool {
    if len(s) < 5 {
        return false
    }
    asciiCount := 0
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
            asciiCount++
        }
    }
    ratio := float64(asciiCount) / float64(len([]rune(s)))
    return ratio > 0.6
}
```

---

## 6. System-Wide Text Processing

### 6.1 Interaction Strategy (Layered)

```
Layer 1 — UI Automation (UIA)     Best: read/write text field directly
Layer 2 — Keyboard Hook + SendInput  Works in most standard apps
Layer 3 — Clipboard replacement    Universal fallback, slightly slower
```

### 6.2 Active Window Detection

```go
// internal/window/detector.go

package window

import (
    "syscall"
    "unsafe"
    "golang.org/x/sys/windows"
)

var (
    user32                       = windows.NewLazySystemDLL("user32.dll")
    procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
    procGetWindowTextW           = user32.NewProc("GetWindowTextW")
    procGetClassNameW            = user32.NewProc("GetClassNameW")
    procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

type WindowInfo struct {
    Handle    windows.HWND
    Title     string
    ClassName string
    ProcessID uint32
    Exe       string
}

func GetForegroundWindowInfo() (*WindowInfo, error) {
    hwnd, _, _ := procGetForegroundWindow.Call()
    if hwnd == 0 {
        return nil, nil
    }

    info := &WindowInfo{Handle: windows.HWND(hwnd)}
    buf := make([]uint16, 512)

    procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
    info.Title = syscall.UTF16ToString(buf)

    procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
    info.ClassName = syscall.UTF16ToString(buf)

    var pid uint32
    procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
    info.ProcessID = pid
    info.Exe = getProcessExe(pid)

    return info, nil
}

type CompatMode int

const (
    CompatModeKeyboard  CompatMode = iota
    CompatModeClipboard
    CompatModeUIA
    CompatModeBlocked
)

var KnownApps = map[string]CompatMode{
    "chrome.exe":      CompatModeKeyboard,
    "firefox.exe":     CompatModeKeyboard,
    "msedge.exe":      CompatModeKeyboard,
    "WINWORD.EXE":     CompatModeKeyboard,
    "notepad.exe":     CompatModeKeyboard,
    "notepad++.exe":   CompatModeKeyboard,
    "Code.exe":        CompatModeKeyboard, // VS Code (Electron)
    "Slack.exe":       CompatModeKeyboard,
    "Discord.exe":     CompatModeKeyboard,
    "WhatsApp.exe":    CompatModeKeyboard,
    "EXCEL.EXE":       CompatModeKeyboard,
    "POWERPNT.EXE":    CompatModeKeyboard,
    "photoshop.exe":   CompatModeClipboard,
    "illustrator.exe": CompatModeClipboard,
    "java.exe":        CompatModeClipboard,
}

func GetCompatMode(exe string) CompatMode {
    if mode, ok := KnownApps[exe]; ok {
        return mode
    }
    return CompatModeKeyboard
}
```

### 6.3 Clipboard Monitor

```go
// internal/hook/clipboard/monitor.go

package clipboard

import (
    "time"
    "unsafe"
    "golang.org/x/sys/windows"
)

var (
    user32               = windows.NewLazySystemDLL("user32.dll")
    procOpenClipboard    = user32.NewProc("OpenClipboard")
    procCloseClipboard   = user32.NewProc("CloseClipboard")
    procGetClipboardData = user32.NewProc("GetClipboardData")
    procSetClipboardData = user32.NewProc("SetClipboardData")
    procEmptyClipboard   = user32.NewProc("EmptyClipboard")
    kernel32             = windows.NewLazySystemDLL("kernel32.dll")
    procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
    procGlobalLock       = kernel32.NewProc("GlobalLock")
    procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
    procGlobalSize       = kernel32.NewProc("GlobalSize")
)

const (
    CF_UNICODETEXT = 13
    GMEM_MOVEABLE  = 0x0002
)

type Monitor struct {
    onChanged func(text string)
    stopCh    chan struct{}
}

func NewMonitor(onChanged func(text string)) *Monitor {
    return &Monitor{onChanged: onChanged, stopCh: make(chan struct{})}
}

func GetClipboardText() (string, error) {
    ret, _, err := procOpenClipboard.Call(0)
    if ret == 0 {
        return "", err
    }
    defer procCloseClipboard.Call()

    handle, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
    if handle == 0 {
        return "", nil
    }

    ptr, _, _ := procGlobalLock.Call(handle)
    if ptr == 0 {
        return "", nil
    }
    defer procGlobalUnlock.Call(handle)

    size, _, _ := procGlobalSize.Call(handle)
    u16 := (*[1 << 20]uint16)(unsafe.Pointer(ptr))[:size/2]
    n := 0
    for n < len(u16) && u16[n] != 0 {
        n++
    }
    return windows.UTF16ToString(u16[:n]), nil
}

func SetClipboardText(text string) error {
    u16, err := windows.UTF16FromString(text)
    if err != nil {
        return err
    }
    size := uintptr(len(u16) * 2)

    h, _, err := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
    if h == 0 {
        return err
    }

    ptr, _, _ := procGlobalLock.Call(h)
    if ptr == 0 {
        return nil
    }
    dst := (*[1 << 20]uint16)(unsafe.Pointer(ptr))
    copy(dst[:], u16)
    procGlobalUnlock.Call(h)

    ret, _, err := procOpenClipboard.Call(0)
    if ret == 0 {
        return err
    }
    defer procCloseClipboard.Call()

    procEmptyClipboard.Call()
    procSetClipboardData.Call(CF_UNICODETEXT, h)
    return nil
}

// Start polls the clipboard every 500ms and calls onChanged when content changes.
func (m *Monitor) Start() {
    go func() {
        var lastText string
        ticker := time.NewTicker(500 * time.Millisecond)
        defer ticker.Stop()
        for {
            select {
            case <-m.stopCh:
                return
            case <-ticker.C:
                text, err := GetClipboardText()
                if err == nil && text != lastText && text != "" {
                    lastText = text
                    m.onChanged(text)
                }
            }
        }
    }()
}

func (m *Monitor) Stop() { close(m.stopCh) }
```

---

## 7. Desktop Application Design

### 7.1 Wails Entry Point

```go
// cmd/main.go

package main

import (
    "embed"
    "log"
    "os"
    "runtime"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/windows"

    "github.com/you/sinhala-assistant/internal/ipc"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    runtime.LockOSThread() // required for Windows message loops

    app := ipc.NewApp()

    err := wails.Run(&options.App{
        Title:             "Sinhala Assistant",
        Width:             420,
        Height:            600,
        Frameless:         true,
        StartHidden:       true,   // tray app — start minimized
        HideWindowOnClose: true,   // hide to tray on X

        AssetServer: &assetserver.Options{Assets: assets},

        Windows: &windows.Options{
            WindowIsTranslucent: true,
            Theme:               windows.Dark,
            CustomTheme: &windows.ThemeSettings{
                DarkModeTitleBar:  windows.RGB(18, 18, 20),
                DarkModeTitleText: windows.RGB(255, 255, 255),
                DarkModeBorder:    windows.RGB(40, 40, 45),
            },
        },

        OnStartup:  app.Startup,
        OnShutdown: app.Shutdown,
        OnDomReady: app.DOMReady,
        Bind:       []interface{}{app},
    })

    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
}
```

### 7.2 IPC Bridge

```go
// internal/ipc/bridge.go

package ipc

import (
    "context"

    "github.com/wailsapp/wails/v2/pkg/runtime"

    "github.com/you/sinhala-assistant/internal/config"
    "github.com/you/sinhala-assistant/internal/engine/normalization"
    "github.com/you/sinhala-assistant/internal/engine/transliteration"
    "github.com/you/sinhala-assistant/internal/hook/clipboard"
    "github.com/you/sinhala-assistant/internal/hook/keyboard"
    "github.com/you/sinhala-assistant/internal/tray"
)

type App struct {
    ctx         context.Context
    cfg         *config.Config
    engine      *transliteration.Engine
    hook        *keyboard.Hook
    processor   *keyboard.Processor
    clipMonitor *clipboard.Monitor
    trayIcon    *tray.Tray
}

func NewApp() *App { return &App{} }

func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
    a.cfg = config.Load()
    a.engine = transliteration.NewEngine(a.cfg.MappingProfile)
    a.hook = keyboard.New()
    a.processor = keyboard.NewProcessor(a.hook, a.engine)
    a.clipMonitor = clipboard.NewMonitor(a.onClipboardChanged)

    a.trayIcon = tray.New(tray.Config{
        OnToggle:   a.ToggleEngine,
        OnSettings: a.ShowSettings,
        OnClipFix:  a.FixClipboard,
        OnQuit:     a.Quit,
    })
    go a.trayIcon.Run()

    if a.cfg.AutoStart {
        a.startEngine()
    }
}

func (a *App) DOMReady(ctx context.Context) {
    runtime.EventsEmit(ctx, "engine:state", a.GetEngineState())
}

func (a *App) Shutdown(ctx context.Context) {
    a.hook.Stop()
    a.clipMonitor.Stop()
    a.trayIcon.Stop()
}

func (a *App) ToggleEngine() bool {
    if a.hook.IsActive() {
        a.hook.SetActive(false)
        runtime.EventsEmit(a.ctx, "engine:stopped", nil)
        a.trayIcon.SetIcon(tray.IconInactive)
        return false
    }
    a.startEngine()
    return true
}

func (a *App) startEngine() {
    if err := a.hook.Start(); err != nil {
        runtime.EventsEmit(a.ctx, "engine:error", err.Error())
        return
    }
    go a.processor.Run()
    go a.hook.RunMessageLoop()
    a.trayIcon.SetIcon(tray.IconActive)
    runtime.EventsEmit(a.ctx, "engine:started", nil)
}

func (a *App) FixClipboard() map[string]interface{} {
    text, err := clipboard.GetClipboardText()
    if err != nil || text == "" {
        return map[string]interface{}{"fixed": false, "error": "empty clipboard"}
    }
    repaired := normalization.RepairString(text)
    if repaired == text {
        return map[string]interface{}{"fixed": false, "message": "no issues found"}
    }
    if err := clipboard.SetClipboardText(repaired); err != nil {
        return map[string]interface{}{"fixed": false, "error": err.Error()}
    }
    return map[string]interface{}{
        "fixed": true, "original": text, "repaired": repaired,
    }
}

func (a *App) AnalyzeText(text string) []normalization.Detection {
    return normalization.Analyze(text)
}

func (a *App) RepairText(text string) string {
    return normalization.RepairString(text)
}

func (a *App) GetEngineState() map[string]interface{} {
    return map[string]interface{}{
        "active":  a.hook != nil && a.hook.IsActive(),
        "mapping": a.cfg.MappingProfile,
        "version": "1.0.0",
    }
}

func (a *App) GetConfig() *config.Config      { return a.cfg }
func (a *App) ShowSettings()                  { runtime.WindowShow(a.ctx) }
func (a *App) Quit()                          { runtime.Quit(a.ctx) }

func (a *App) SaveConfig(cfg config.Config) error {
    a.cfg = &cfg
    return config.Save(&cfg)
}

func (a *App) onClipboardChanged(text string) {
    if !a.cfg.AutoFixClipboard {
        return
    }
    issues := normalization.Analyze(text)
    if len(issues) > 0 {
        runtime.EventsEmit(a.ctx, "clipboard:issues", map[string]interface{}{
            "count": len(issues), "text": text,
        })
    }
}
```

### 7.3 System Tray

```go
// internal/tray/tray.go

package tray

import "github.com/getlantern/systray"

type IconState int

const (
    IconActive   IconState = iota
    IconInactive
    IconError
)

type Config struct {
    OnToggle   func()
    OnSettings func()
    OnClipFix  func()
    OnQuit     func()
}

type Tray struct {
    cfg        Config
    stopCh     chan struct{}
    iconCh     chan IconState
    menuToggle *systray.MenuItem
}

func New(cfg Config) *Tray {
    return &Tray{cfg: cfg, stopCh: make(chan struct{}), iconCh: make(chan IconState, 1)}
}

func (t *Tray) Run()  { systray.Run(t.onReady, t.onExit) }
func (t *Tray) Stop() { systray.Quit() }

func (t *Tray) onReady() {
    systray.SetIcon(iconInactive)
    systray.SetTooltip("Sinhala Assistant — Inactive")

    t.menuToggle = systray.AddMenuItem("Enable Sinhala Typing", "Toggle the keyboard engine")
    mSettings   := systray.AddMenuItem("Settings", "Open settings panel")
    mClipFix    := systray.AddMenuItem("Fix Clipboard", "Repair Sinhala in clipboard")
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Exit Sinhala Assistant")

    go func() {
        for {
            select {
            case <-t.menuToggle.ClickedCh: t.cfg.OnToggle()
            case <-mSettings.ClickedCh:   t.cfg.OnSettings()
            case <-mClipFix.ClickedCh:    t.cfg.OnClipFix()
            case <-mQuit.ClickedCh:       t.cfg.OnQuit()
            case state := <-t.iconCh:     t.applyIcon(state)
            case <-t.stopCh:              return
            }
        }
    }()
}

func (t *Tray) SetIcon(state IconState) {
    select {
    case t.iconCh <- state:
    default:
    }
}

func (t *Tray) applyIcon(state IconState) {
    switch state {
    case IconActive:
        systray.SetIcon(iconActive)
        systray.SetTooltip("Sinhala Assistant — Active")
        t.menuToggle.SetTitle("Disable Sinhala Typing")
    case IconInactive:
        systray.SetIcon(iconInactive)
        systray.SetTooltip("Sinhala Assistant — Inactive")
        t.menuToggle.SetTitle("Enable Sinhala Typing")
    case IconError:
        systray.SetIcon(iconError)
        systray.SetTooltip("Sinhala Assistant — Error")
    }
}

func (t *Tray) onExit() {}
```

### 7.4 React Frontend Structure

```
frontend/src/
├── App.tsx
├── components/
│   ├── TrayPanel/
│   │   ├── TrayPanel.tsx       # Main panel when tray clicked
│   │   ├── StatusCard.tsx      # Big toggle + status indicator
│   │   └── QuickActions.tsx    # Fix clipboard, switch mode buttons
│   │
│   ├── Settings/
│   │   ├── SettingsPanel.tsx
│   │   ├── GeneralSettings.tsx
│   │   ├── MappingSettings.tsx     # Switch keyboard layout
│   │   ├── HotkeySettings.tsx
│   │   └── AppCompatibility.tsx    # Per-app whitelist manager
│   │
│   ├── Analyzer/
│   │   ├── TextAnalyzer.tsx    # Paste + analyze Sinhala text
│   │   ├── IssueList.tsx       # Show detected issues
│   │   └── DiffView.tsx        # Original vs repaired diff
│   │
│   └── KeyboardLayout/
│       ├── KeyLayout.tsx        # Visual keyboard with Sinhala labels
│       └── KeyCap.tsx
│
├── hooks/
│   ├── useEngine.ts
│   └── useWails.ts
│
└── store/
    ├── engineStore.ts           # Zustand — engine on/off, mapping mode
    └── settingsStore.ts
```

**Engine hook:**

```typescript
// frontend/src/hooks/useEngine.ts

import { useEffect } from 'react'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import { useEngineStore } from '../store/engineStore'

export function useEngine() {
  const { setActive, setError } = useEngineStore()

  useEffect(() => {
    EventsOn('engine:started', () => setActive(true))
    EventsOn('engine:stopped', () => setActive(false))
    EventsOn('engine:error', (msg: string) => setError(msg))

    return () => {
      EventsOff('engine:started')
      EventsOff('engine:stopped')
      EventsOff('engine:error')
    }
  }, [])
}
```

**Engine store:**

```typescript
// frontend/src/store/engineStore.ts

import { create } from 'zustand'

interface EngineState {
  active: boolean
  mapping: string
  error: string | null
  setActive: (v: boolean) => void
  setMapping: (v: string) => void
  setError: (v: string | null) => void
}

export const useEngineStore = create<EngineState>((set) => ({
  active:   false,
  mapping:  'phonetic',
  error:    null,
  setActive:  (active)  => set({ active, error: null }),
  setMapping: (mapping) => set({ mapping }),
  setError:   (error)   => set({ error }),
}))
```

---

## 8. Performance Optimization

### 8.1 Hook Performance Rules

The `WH_KEYBOARD_LL` callback has a **200ms hard timeout** — Windows auto-removes hooks that exceed it.

| Rule | Detail |
|---|---|
| Zero allocation in hook callback | Pre-allocate all buffers at startup |
| No blocking calls in callback | Use buffered channel, drop if full |
| Transliteration on separate goroutine | Hook callback only enqueues |
| UI updates debounced | Never from hook goroutine |
| Pre-compiled trie | Build at startup, read-only in hot path |

### 8.2 Pre-allocated Ring Buffer

```go
// internal/hook/keyboard/buffer.go

package keyboard

import "sync"

type InputBuffer struct {
    buf  [512]KeyEvent
    head uint32
    tail uint32
    mu   sync.Mutex
}

func (b *InputBuffer) Push(e KeyEvent) bool {
    b.mu.Lock()
    next := (b.tail + 1) % uint32(len(b.buf))
    if next == b.head {
        b.mu.Unlock()
        return false
    }
    b.buf[b.tail] = e
    b.tail = next
    b.mu.Unlock()
    return true
}

func (b *InputBuffer) Pop() (KeyEvent, bool) {
    b.mu.Lock()
    defer b.mu.Unlock()
    if b.head == b.tail {
        return KeyEvent{}, false
    }
    e := b.buf[b.head]
    b.head = (b.head + 1) % uint32(len(b.buf))
    return e, true
}
```

---

## 9. Security & Stability

### 9.1 Hook Protection

```go
// Watchdog reinstalls hook if Windows silently drops it
func (h *Hook) WatchdogRoutine(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for range ticker.C {
        if !h.active.Load() {
            continue
        }
        if h.handle == 0 {
            h.Stop()
            h.Start()
        }
    }
}
```

### 9.2 Injection Rate Limiter

```go
type InjectionGuard struct {
    lastInject time.Time
    count      int
    mu         sync.Mutex
}

const maxInjectionsPerSecond = 50

func (g *InjectionGuard) Allow() bool {
    g.mu.Lock()
    defer g.mu.Unlock()
    now := time.Now()
    if now.Sub(g.lastInject) > time.Second {
        g.count = 0
        g.lastInject = now
    }
    if g.count >= maxInjectionsPerSecond {
        return false
    }
    g.count++
    return true
}
```

### 9.3 Windows Compatibility Matrix

| Feature | Win 7/8 | Win 10 | Win 11 |
|---|---|---|---|
| WH_KEYBOARD_LL | ✓ | ✓ | ✓ |
| SendInput Unicode | ✓ | ✓ | ✓ |
| UI Automation | Partial | Full | Full |
| Tray Icon | ✓ | ✓ | ✓ |
| Toast Notifications | ✗ | ✓ | ✓ |

> **UAC Note**: A non-elevated hook cannot intercept keystrokes in an elevated (admin) process. Fall back to clipboard mode for admin windows.

---

## 10. MVP Development Roadmap

### Phase 1 — Core Engine (Weeks 1–3)

```
Week 1: Unicode research + mapping tables
        • Complete phonetic mapping JSON with all 200+ combinations
        • Complete Wijesekera mapping JSON
        • Unit tests: kombuwa, ispili, papili, rakar, repaya specifically

Week 2: Transliteration engine
        • Trie + state machine implementation
        • All vowel sign combinations
        • Consonant clusters with ZWJ
        • 50+ word test corpus

Week 3: Normalization engine
        • Malformed Sinhala detector
        • Repair engine (reordering rules)
        • Wijesekera legacy converter
        • CLI tool: sinhala-fix.exe
```

**Milestone**: `sinhala-fix.exe` CLI that takes text and outputs repaired Sinhala.

### Phase 2 — Wails App + Tray (Weeks 4–6)

```
Week 4: Wails scaffold
        • Project structure
        • React + Tailwind setup
        • IPC bridge skeleton
        • System tray (on/off icon only)

Week 5: Settings UI
        • Toggle panel
        • Mapping switcher
        • Clipboard fix button
        • JSON config persistence

Week 6: Text analyzer UI
        • Paste → analyze → show issues
        • Original vs repaired diff view
        • Copy repaired to clipboard
```

**Milestone**: Tray app with working clipboard fix.

### Phase 3 — Keyboard Hook (Weeks 7–10)

```
Week 7: WH_KEYBOARD_LL hook
        • SetWindowsHookEx implementation
        • LLKHF_INJECTED guard
        • Message loop thread

Week 8: Text injection
        • SendInput Unicode
        • Backspace + replace flow
        • Test: Notepad, Chrome, Firefox

Week 9: App compatibility
        • Active window detection
        • Per-app compat mode
        • Test: Word, VS Code, Slack, Discord

Week 10: Stability
        • Watchdog goroutine
        • Hook recovery on drop
        • Elevated process fallback
        • Performance profiling
```

**Milestone**: Real-time Sinhala typing working globally.

### Phase 4 — Polish & Release (Weeks 11–13)

```
Week 11: Hotkey system
         • Global toggle hotkey (Ctrl+Alt+S)
         • Hotkey configuration UI

Week 12: Installer
         • NSIS installer script
         • Windows startup registry entry

Week 13: Testing & release
         • Clean-install test on Win 10/11
         • App compatibility matrix verified
         • GitHub release
```

---

## 11. Mapping & Database Design

### 11.1 Mapping JSON Structure

```json
{
  "meta": {
    "id": "phonetic-v1",
    "name": "Sinhala Phonetic",
    "version": "1.0.0",
    "description": "Phonetic transliteration based on English pronunciation"
  },

  "settings": {
    "use_zwj_clusters": true,
    "auto_fix_kombuwa": true,
    "commit_on_space": true,
    "commit_on_punctuation": true
  },

  "independent_vowels": {
    "a":  {"unicode": "U+0D85", "sinhala": "අ"},
    "aa": {"unicode": "U+0D86", "sinhala": "ආ"},
    "ae": {"unicode": "U+0D87", "sinhala": "ඇ"},
    "i":  {"unicode": "U+0D89", "sinhala": "ඉ"},
    "ii": {"unicode": "U+0D8A", "sinhala": "ඊ"},
    "u":  {"unicode": "U+0D8B", "sinhala": "උ"},
    "uu": {"unicode": "U+0D8C", "sinhala": "ඌ"},
    "e":  {"unicode": "U+0D91", "sinhala": "එ"},
    "ee": {"unicode": "U+0D92", "sinhala": "ඒ"},
    "o":  {"unicode": "U+0D94", "sinhala": "ඔ"},
    "oo": {"unicode": "U+0D95", "sinhala": "ඕ"},
    "au": {"unicode": "U+0D96", "sinhala": "ඖ"}
  },

  "consonants": {
    "k":  {"unicode": "U+0D9A", "sinhala": "ක", "category": "velar"},
    "kh": {"unicode": "U+0D9B", "sinhala": "ඛ", "category": "velar_aspirated"},
    "g":  {"unicode": "U+0D9C", "sinhala": "ග", "category": "velar"},
    "gh": {"unicode": "U+0D9D", "sinhala": "ඝ", "category": "velar_aspirated"},
    "ng": {"unicode": "U+0D9E", "sinhala": "ඞ", "category": "velar_nasal"},
    "c":  {"unicode": "U+0DA0", "sinhala": "ච", "category": "palatal"},
    "j":  {"unicode": "U+0DA2", "sinhala": "ජ", "category": "palatal"},
    "jh": {"unicode": "U+0DA3", "sinhala": "ඣ", "category": "palatal_aspirated"},
    "ny": {"unicode": "U+0DA4", "sinhala": "ඤ", "category": "palatal_nasal"},
    "T":  {"unicode": "U+0DA7", "sinhala": "ට", "category": "retroflex"},
    "Th": {"unicode": "U+0DA8", "sinhala": "ඨ", "category": "retroflex_aspirated"},
    "D":  {"unicode": "U+0DA9", "sinhala": "ඩ", "category": "retroflex"},
    "Dh": {"unicode": "U+0DAA", "sinhala": "ඪ", "category": "retroflex_aspirated"},
    "N":  {"unicode": "U+0DAB", "sinhala": "ණ", "category": "retroflex_nasal"},
    "t":  {"unicode": "U+0DAD", "sinhala": "ත", "category": "dental"},
    "th": {"unicode": "U+0DAE", "sinhala": "ථ", "category": "dental_aspirated"},
    "d":  {"unicode": "U+0DAF", "sinhala": "ද", "category": "dental"},
    "dh": {"unicode": "U+0DB0", "sinhala": "ධ", "category": "dental_aspirated"},
    "n":  {"unicode": "U+0DB1", "sinhala": "න", "category": "dental_nasal"},
    "p":  {"unicode": "U+0DB4", "sinhala": "ප", "category": "bilabial"},
    "ph": {"unicode": "U+0DB5", "sinhala": "ඵ", "category": "bilabial_aspirated"},
    "b":  {"unicode": "U+0DB6", "sinhala": "බ", "category": "bilabial"},
    "bh": {"unicode": "U+0DB7", "sinhala": "භ", "category": "bilabial_aspirated"},
    "m":  {"unicode": "U+0DB8", "sinhala": "ම", "category": "bilabial_nasal"},
    "mb": {"unicode": "U+0DB9", "sinhala": "ඹ", "category": "bilabial_nasal"},
    "y":  {"unicode": "U+0DBA", "sinhala": "ය", "category": "approximant"},
    "r":  {"unicode": "U+0DBB", "sinhala": "ර", "category": "trill"},
    "l":  {"unicode": "U+0DBD", "sinhala": "ල", "category": "lateral"},
    "v":  {"unicode": "U+0DC0", "sinhala": "ව", "category": "labiodental"},
    "sh": {"unicode": "U+0DC1", "sinhala": "ශ", "category": "sibilant"},
    "Sh": {"unicode": "U+0DC2", "sinhala": "ෂ", "category": "retroflex_sibilant"},
    "s":  {"unicode": "U+0DC3", "sinhala": "ස", "category": "sibilant"},
    "h":  {"unicode": "U+0DC4", "sinhala": "හ", "category": "glottal"},
    "L":  {"unicode": "U+0DC5", "sinhala": "ළ", "category": "retroflex_lateral"},
    "f":  {"unicode": "U+0DC6", "sinhala": "ෆ", "category": "labiodental"}
  },

  "vowel_signs": {
    "a":  {"unicode": "U+0DCF", "sinhala": "ා", "name": "aa_sign"},
    "i":  {"unicode": "U+0DD2", "sinhala": "ි", "name": "ispili"},
    "I":  {"unicode": "U+0DD3", "sinhala": "ී", "name": "papili"},
    "u":  {"unicode": "U+0DD4", "sinhala": "ු", "name": "u_sign"},
    "U":  {"unicode": "U+0DD6", "sinhala": "ූ", "name": "uu_sign"},
    "E":  {"unicode": "U+0DD9", "sinhala": "ෙ", "name": "kombuwa", "prebase": true},
    "ee": {"unicode": "U+0DDA", "sinhala": "ේ", "name": "deerga_kombuwa"},
    "o":  {
      "sequence": ["U+0DD9", "U+0DCF"],
      "sinhala": "ො",
      "name": "o_sign",
      "note": "kombuwa + aa sign sequence"
    },
    "oo": {"sequence": ["U+0DD9", "U+0DDD"], "sinhala": "ෝ", "name": "oo_sign"}
  },

  "special": {
    "H":  {"unicode": "U+0DCA", "sinhala": "්", "name": "al_lakuna"},
    "^":  {"unicode": "U+0DCA", "sinhala": "්", "name": "al_lakuna"},
    "M":  {"unicode": "U+0D82", "sinhala": "ං", "name": "anusvaraya"},
    "H2": {"unicode": "U+0D83", "sinhala": "ඃ", "name": "visargaya"}
  },

  "clusters": {
    "rakar": {
      "pattern": "consonant + al_lakuna + r",
      "unicode_sequence": "consonant + U+0DCA + U+200D + U+0DBB",
      "example": "kHr → ක්‍ර"
    },
    "yansaya": {
      "pattern": "consonant + al_lakuna + y",
      "unicode_sequence": "consonant + U+0DCA + U+200D + U+0DBA",
      "example": "kHy → ක්‍ය"
    },
    "repaya": {
      "pattern": "r + al_lakuna + consonant",
      "unicode_sequence": "U+0DBB + U+0DCA + U+200D + consonant",
      "example": "rHk → ර්‍ක"
    }
  }
}
```

---

## 12. UI/UX Design

### 12.1 Design System

| Token | Value | Usage |
|---|---|---|
| `bg-primary` | `#0F0F12` | App background |
| `bg-card` | `#1A1A20` | Card backgrounds |
| `bg-hover` | `#242430` | Hover/active states |
| `accent-blue` | `#4F8EF7` | Primary accent |
| `accent-green` | `#3ECF8E` | Active/success state |
| `accent-red` | `#E5534B` | Error/disabled |
| `text-primary` | `#F0F0F5` | Main text |
| `text-muted` | `#8888AA` | Secondary text |
| `border` | `#2A2A38` | Borders |

### 12.2 Main Tray Panel

```
┌──────────────────────────────────────┐
│  ◉ Sinhala Assistant          —  ╳  │
├──────────────────────────────────────┤
│                                      │
│   ┌────────────────────────────┐    │
│   │                            │    │
│   │   ●  SINHALA TYPING        │    │
│   │      ACTIVE                │    │
│   │                      ●─○  │    │
│   └────────────────────────────┘    │
│                                      │
│   Layout: [ Phonetic ▾ ]            │
│                                      │
│   ┌──────────┐  ┌──────────────┐   │
│   │ Fix Clip │  │ Analyze Text │   │
│   └──────────┘  └──────────────┘   │
│                                      │
│  ─────────────────────────────────  │
│                                      │
│   Recent Fixes                       │
│   • 3 issues repaired (2 min ago)   │
│   • Kombuwa reordered in clipboard  │
│                                      │
│  ─────────────────────────────────  │
│  [⚙ Settings]  [?]  [Quit]         │
└──────────────────────────────────────┘
```

### 12.3 Text Analyzer Panel

```
┌──────────────────────────────────────────────────┐
│  Text Analyzer                             ← Back │
├──────────────────────────────────────────────────┤
│                                                    │
│  Paste Sinhala text:                              │
│  ┌──────────────────────────────────────────┐    │
│  │ ෙකොළඹ  (broken — kombuwa before base)    │    │
│  └──────────────────────────────────────────┘    │
│                                  [Analyze ▶]      │
│                                                    │
│  Issues Found: 2                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ ⚠  Position 0: Kombuwa before base       │    │
│  │    Found: ෙ + ක  →  Should be: ක + ෙ    │    │
│  │                                          │    │
│  │ ⚠  Position 5: Missing ZWJ in rakar     │    │
│  │    Found: ්ර  →  Should be: ්‍ර          │    │
│  └──────────────────────────────────────────┘    │
│                                                    │
│  Repaired:                                        │
│  ┌──────────────────────────────────────────┐    │
│  │ කොළඹ  ✓                                 │    │
│  └──────────────────────────────────────────┘    │
│                                                    │
│  [Copy Repaired]              [Replace Clipboard] │
└──────────────────────────────────────────────────┘
```

### 12.4 Visual Keyboard Component

```typescript
// frontend/src/components/KeyboardLayout/KeyLayout.tsx

const KEYBOARD_ROWS = [
  ['q','w','e','r','t','y','u','i','o','p'],
  ['a','s','d','f','g','h','j','k','l'],
  ['z','x','c','v','b','n','m'],
]

export function KeyLayout() {
  const { mapping } = useEngineStore()

  return (
    <div className="p-4 bg-[#1A1A20] rounded-xl">
      <div className="text-xs text-[#8888AA] mb-3">
        Keyboard Layout — {mapping}
      </div>
      {KEYBOARD_ROWS.map((row, i) => (
        <div key={i} className="flex gap-1 mb-1 justify-center">
          {row.map((key) => {
            const entry = PhoneticMap[key]
            return <KeyCap key={key} english={key} sinhala={entry?.sinhala} />
          })}
        </div>
      ))}
    </div>
  )
}

function KeyCap({ english, sinhala }: { english: string; sinhala?: string }) {
  return (
    <div className={`
      w-9 h-9 rounded-md border flex flex-col items-center justify-center
      ${sinhala
        ? 'bg-[#242430] border-[#4F8EF7]/40 text-white'
        : 'bg-[#1A1A20] border-[#2A2A38] text-[#4A4A5A]'
      }
    `}>
      {sinhala && <span className="text-[10px] text-[#4F8EF7]">{sinhala}</span>}
      <span className="text-[9px] text-[#8888AA]">{english}</span>
    </div>
  )
}
```

---

## 13. Future AI Features

### 13.1 Predictive Typing — N-gram Model

```go
// internal/engine/prediction/predictor.go

package prediction

import (
    "sort"
    "strings"
)

type NGramModel struct {
    bigrams  map[string]map[string]int
    unigrams map[string]int
    total    int
}

type Suggestion struct {
    Word  string
    Score float64
}

func (m *NGramModel) Predict(prefix, prevWord string, topN int) []Suggestion {
    candidates := make(map[string]float64)

    for word, count := range m.unigrams {
        if strings.HasPrefix(word, prefix) {
            candidates[word] = float64(count) / float64(m.total)
        }
    }

    if prevWord != "" {
        if nextWords, ok := m.bigrams[prevWord]; ok {
            for word, count := range nextWords {
                if strings.HasPrefix(word, prefix) {
                    candidates[word] += float64(count) * 2.0
                }
            }
        }
    }

    suggestions := make([]Suggestion, 0, len(candidates))
    for word, score := range candidates {
        suggestions = append(suggestions, Suggestion{Word: word, Score: score})
    }
    sort.Slice(suggestions, func(i, j int) bool {
        return suggestions[i].Score > suggestions[j].Score
    })

    if len(suggestions) > topN {
        return suggestions[:topN]
    }
    return suggestions
}
```

### 13.2 Claude API Integration

```go
// internal/engine/ai/corrector.go

package ai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type SinhalaCorrector struct {
    apiKey string
    client *http.Client
}

func New(apiKey string) *SinhalaCorrector {
    return &SinhalaCorrector{
        apiKey: apiKey,
        client: &http.Client{Timeout: 3 * time.Second},
    }
}

// CorrectSinhala uses Claude Haiku to fix Sinhala spelling/grammar.
// Called asynchronously — never blocks the typing pipeline.
func (c *SinhalaCorrector) CorrectSinhala(ctx context.Context, text string) (string, error) {
    body, _ := json.Marshal(map[string]interface{}{
        "model":      "claude-haiku-4-5-20251001",
        "max_tokens": 256,
        "messages": []map[string]string{
            {
                "role": "user",
                "content": fmt.Sprintf(
                    "Fix any Sinhala spelling or Unicode ordering issues in this text. "+
                        "Return ONLY the corrected Sinhala text:\n\n%s", text,
                ),
            },
        },
    })

    req, err := http.NewRequestWithContext(ctx, "POST",
        "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
    if err != nil {
        return text, err
    }
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return text, err
    }
    defer resp.Body.Close()

    var result struct {
        Content []struct{ Text string `json:"text"` } `json:"content"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return text, err
    }
    if len(result.Content) == 0 {
        return text, nil
    }
    return result.Content[0].Text, nil
}
```

### 13.3 Legacy OCR Cleanup Pipeline

```
Scanned Document → Tesseract OCR (Sinhala traineddata)
                         │
                         ▼
                Raw OCR Output (likely broken Unicode)
                         │
                         ▼
                normalization.Repair()    ← rule-based engine
                         │
                         ▼
                ai.CorrectSinhala()       ← Claude for grammar/spelling
                         │
                         ▼
                Clean Sinhala Unicode Output
```

---

## 14. Development Best Practices

### 14.1 go.mod Dependencies

```
module github.com/you/sinhala-assistant

go 1.22

require (
    github.com/wailsapp/wails/v2   v2.9.1
    golang.org/x/sys               v0.21.0   // Windows API
    golang.org/x/text              v0.16.0   // Unicode normalization
    github.com/getlantern/systray  v1.2.2    // System tray
    github.com/spf13/viper         v1.19.0   // Config management
    go.uber.org/zap                v1.27.0   // Structured logging
    github.com/stretchr/testify    v1.9.0    // Testing
)
```

### 14.2 Unit Tests

```go
// internal/engine/transliteration/engine_test.go

package transliteration_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/you/sinhala-assistant/internal/engine/transliteration"
)

func TestKombuwaOrdering(t *testing.T) {
    e := transliteration.NewEngine("phonetic")
    result, _ := e.Feed("kE")
    runes := []rune(result.Output)
    assert.Equal(t, rune(0x0D9A), runes[0], "consonant must precede kombuwa")
    assert.Equal(t, rune(0x0DD9), runes[1], "kombuwa must follow consonant")
}

func TestIspiliAttachment(t *testing.T) {
    e := transliteration.NewEngine("phonetic")
    result, _ := e.Feed("ki")
    assert.Equal(t, "කි", result.Output)
}

func TestRakarZWJ(t *testing.T) {
    e := transliteration.NewEngine("phonetic")
    e.Feed("k")
    e.Feed("H")
    result, _ := e.Feed("r")
    runes := []rune(result.Output)
    assert.Equal(t, rune(0x0D9A), runes[0]) // ක
    assert.Equal(t, rune(0x0DCA), runes[1]) // ්
    assert.Equal(t, rune(0x200D), runes[2]) // ZWJ
    assert.Equal(t, rune(0x0DBB), runes[3]) // ර
}

func TestRepayaOrdering(t *testing.T) {
    e := transliteration.NewEngine("phonetic")
    e.Feed("r")
    e.Feed("H")
    result, _ := e.Feed("k")
    runes := []rune(result.Output)
    assert.Equal(t, rune(0x0DBB), runes[0]) // ර
    assert.Equal(t, rune(0x0DCA), runes[1]) // ්
    assert.Equal(t, rune(0x200D), runes[2]) // ZWJ
    assert.Equal(t, rune(0x0D9A), runes[3]) // ක
}

func TestVowelSigns(t *testing.T) {
    cases := []struct{ input, expected string }{
        {"ka",  "කා"},
        {"ki",  "කි"},
        {"kI",  "කී"},
        {"ku",  "කු"},
        {"kU",  "කූ"},
        {"kE",  "කෙ"},
        {"ko",  "කො"},
        {"koo", "කෝ"},
    }
    for _, tc := range cases {
        t.Run(tc.input, func(t *testing.T) {
            e := transliteration.NewEngine("phonetic")
            var result transliteration.TransliterationResult
            for _, ch := range tc.input {
                result, _ = e.Feed(string(ch))
            }
            assert.Equal(t, tc.expected, result.Output)
        })
    }
}
```

```go
// internal/engine/normalization/repair_test.go

package normalization_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/you/sinhala-assistant/internal/engine/normalization"
)

func TestKombuwaRepair(t *testing.T) {
    broken := string([]rune{0x0DD9, 0x0D9A}) // ෙ + ක (wrong order)
    repaired := normalization.Repair(broken)
    runes := []rune(repaired)
    assert.Equal(t, rune(0x0D9A), runes[0])
    assert.Equal(t, rune(0x0DD9), runes[1])
}

func TestDoubleViramaRepair(t *testing.T) {
    broken := string([]rune{0x0D9A, 0x0DCA, 0x0DCA, 0x0DC3}) // ක + ්් + ස
    repaired := normalization.Repair(broken)
    assert.Len(t, normalization.Analyze(repaired), 0)
}

func TestWijeskeeraConversion(t *testing.T) {
    converted := normalization.ConvertLegacyWijesekera("lv")
    assert.Equal(t, "ලව", converted)
}
```

### 14.3 Config Schema

```go
// internal/config/config.go

package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    MappingProfile   string            `json:"mapping_profile"`
    AutoStart        bool              `json:"auto_start"`
    StartWithWindows bool              `json:"start_with_windows"`
    AutoFixClipboard bool              `json:"auto_fix_clipboard"`
    UseZWJClusters   bool              `json:"use_zwj_clusters"`
    ToggleHotkey     string            `json:"toggle_hotkey"`
    AppOverrides     map[string]string `json:"app_overrides"`
    Theme            string            `json:"theme"`
    LogLevel         string            `json:"log_level"`
}

func DefaultConfig() *Config {
    return &Config{
        MappingProfile: "phonetic",
        UseZWJClusters: true,
        ToggleHotkey:   "ctrl+alt+s",
        AppOverrides:   map[string]string{},
        Theme:          "dark",
        LogLevel:       "info",
    }
}

func configPath() string {
    return filepath.Join(os.Getenv("APPDATA"), "SinhalaAssistant", "config.json")
}

func Load() *Config {
    cfg := DefaultConfig()
    data, err := os.ReadFile(configPath())
    if err != nil {
        return cfg
    }
    json.Unmarshal(data, cfg)
    return cfg
}

func Save(cfg *Config) error {
    path := configPath()
    os.MkdirAll(filepath.Dir(path), 0755)
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

### 14.4 Windows Startup Registry

```go
// internal/config/startup.go

package config

import "golang.org/x/sys/windows/registry"

const (
    startupKey = `Software\Microsoft\Windows\CurrentVersion\Run`
    appName    = "SinhalaAssistant"
)

func SetStartWithWindows(enable bool, exePath string) error {
    k, err := registry.OpenKey(registry.CURRENT_USER, startupKey, registry.SET_VALUE)
    if err != nil {
        return err
    }
    defer k.Close()

    if !enable {
        return k.DeleteValue(appName)
    }
    return k.SetStringValue(appName, `"`+exePath+`" --tray`)
}
```

### 14.5 Build Script

```powershell
# build.ps1

param(
    [string]$Version = "1.0.0",
    [switch]$Release
)

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "Building Sinhala Assistant v$Version..."

Set-Location frontend
npm install
npm run build
Set-Location ..

if ($Release) {
    wails build -platform windows/amd64 -ldflags "-X main.Version=$Version" -nsis
} else {
    wails build -platform windows/amd64 -debug
}

Write-Host "Done: build/bin/sinhala-assistant.exe"
```

### 14.6 NSIS Installer

```nsis
; installer/sinhala-assistant.nsi

!define APP_NAME     "Sinhala Assistant"
!define APP_VERSION  "1.0.0"
!define APP_EXE      "sinhala-assistant.exe"

Name "${APP_NAME}"
OutFile "SinhalaAssistant-Setup-${APP_VERSION}.exe"
InstallDir "$PROGRAMFILES64\${APP_NAME}"
RequestExecutionLevel user

!include "MUI2.nsh"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Main"
    SetOutPath "$INSTDIR"
    File "${APP_EXE}"
    File /r "assets\"

    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
    CreateShortcut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"

    WriteUninstaller "$INSTDIR\Uninstall.exe"

    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
        "DisplayName" "${APP_NAME}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
        "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" \
        "DisplayVersion" "${APP_VERSION}"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\${APP_EXE}"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir /r "$INSTDIR"
    Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
    Delete "$DESKTOP\${APP_NAME}.lnk"
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
SectionEnd
```

---

## Sinhala Unicode Quick Reference

```
INDEPENDENT VOWELS              VOWEL SIGNS (after consonant)
U+0D85  අ   a                  U+0DCF  ා    aa
U+0D86  ආ   aa                 U+0DD2  ි    i   (ispili)
U+0D87  ඇ   ae                 U+0DD3  ී    ii  (papili)
U+0D89  ඉ   i                  U+0DD4  ු    u
U+0D8A  ඊ   ii                 U+0DD6  ූ    uu
U+0D8B  උ   u                  U+0DD9  ෙ    e   (kombuwa — PREBASE)
U+0D8C  ඌ   uu                 U+0DDA  ේ    ee
U+0D91  එ   e                  U+0DDC  ො    o   (= ෙ + ා)
U+0D92  ඒ   ee                 U+0DDD  ෝ    oo
U+0D94  ඔ   o                  U+0DDE  ෞ    au
U+0D95  ඕ   oo

SPECIAL CHARACTERS
U+0DCA  ්    Al-lakuna (virama / hal kirima)
U+0D82  ං    Anusvaraya
U+0D83  ඃ    Visargaya
U+200D  ZWJ  — forces touching/conjunct cluster forms
U+200C  ZWNJ — forces non-touching forms

CLUSTER SEQUENCES
Rakar:   consonant + ් + ZWJ + ර  → subscript ra below consonant
Yansaya: consonant + ් + ZWJ + ය  → subscript ya below consonant
Repaya:  ර + ් + ZWJ + consonant  → superscript ra above consonant
```

---

## Implementation Priority

Build in this exact order — each layer depends on the previous:

1. **Mapping JSON** — complete character tables with unit tests
2. **Normalization engine** — pure Go, fully tested, zero Windows dependency
3. **Transliteration state machine** — test every Sinhala syllable type
4. **Clipboard fixer CLI** — proves the engine before any UI
5. **Wails scaffold + tray** — clipboard fix button only
6. **Keyboard hook** — add last; hardest to debug
7. **App compatibility layer** — tune per-app after hook works in Notepad
8. **AI features** — after core is stable and shipped

> The most critical pre-requisite: build a test corpus of 50 Sinhala words covering every combination (kombuwa, rakar, repaya, yansaya, ispili, papili) and verify your engine produces bit-perfect Unicode against a known-good reference before writing any Windows API code.
