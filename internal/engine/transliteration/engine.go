package transliteration

// Config holds transliteration behaviour settings.
type Config struct {
	UseZWJClusters      bool
	AutoFixKombuwa      bool
	CommitOnSpace       bool
	CommitOnPunctuation bool
}

// MappingTable is the interface the state machine uses for key lookups.
type MappingTable interface {
	Lookup(raw string) (Token, bool)
}

// PhoneticMappingTable wraps PhoneticMap via a trie.
type PhoneticMappingTable struct {
	trie *Trie
}

func NewPhoneticMappingTable() *PhoneticMappingTable {
	t := NewTrie()
	for key, entry := range PhoneticMap {
		t.Insert(key, entry)
	}
	return &PhoneticMappingTable{trie: t}
}

func (p *PhoneticMappingTable) Lookup(raw string) (Token, bool) {
	entry, _, _ := p.trie.LookupWithBuffer(raw)
	if entry == nil {
		return Token{}, false
	}
	return Token{Type: entry.Type, Sinhala: entry.Output, Raw: raw}, true
}

// Engine wraps a StateMachine and performs prefix-aware digraph resolution.
// Feed must be called with one raw Latin letter at a time.
// Non-letter keys should trigger Flush before being passed to the application.
type Engine struct {
	sm        *StateMachine
	trie      *Trie
	prefixBuf string // raw chars pending trie resolution (not yet injected)
}

func NewEngine(profile string, useZWJ bool) *Engine {
	cfg := Config{
		UseZWJClusters:      useZWJ,
		AutoFixKombuwa:      true,
		CommitOnSpace:       true,
		CommitOnPunctuation: true,
	}
	pmt := NewPhoneticMappingTable()
	return &Engine{
		sm:   NewStateMachine(pmt, cfg),
		trie: pmt.trie,
	}
}

// Feed processes a single intercepted Latin character.
// Returns (result, true) when the SM produced injectable output;
// (zero, false) when still buffering for a longer digraph match.
func (e *Engine) Feed(char string) (TransliterationResult, bool) {
	e.prefixBuf += char
	return e.resolvePrefix()
}

// resolvePrefix attempts to match e.prefixBuf against the trie and feed a token
// to the SM when a definitive match is found.
func (e *Engine) resolvePrefix() (TransliterationResult, bool) {
	entry, matchLen, isPrefix := e.trie.LookupWithBuffer(e.prefixBuf)

	if isPrefix {
		// Current buffer could be the start of a longer key (e.g. "k" before "kh").
		// Don't inject yet — wait for the next character.
		return TransliterationResult{}, false
	}

	if entry == nil {
		// Nothing matched — commit any SM state and pass the raw chars through
		// as Unicode so unmapped keys (e.g. 'x', 'q') remain visible.
		raw := e.prefixBuf
		e.prefixBuf = ""
		smResult := e.sm.commit()
		e.sm.reset()
		return TransliterationResult{
			Output:      smResult.Output + raw,
			DeleteCount: smResult.DeleteCount,
			Committed:   true,
		}, true
	}

	// Definitive match: feed the matched token, recurse on any trailing chars.
	matched := e.prefixBuf[:matchLen]
	remaining := e.prefixBuf[matchLen:]
	e.prefixBuf = ""

	token := Token{Type: entry.Type, Sinhala: entry.Output, Raw: matched}
	result, ok := e.sm.feedToken(token)
	if result.Committed {
		e.sm.reset()
	}

	if remaining != "" {
		e.prefixBuf = remaining
		r2, ok2 := e.resolvePrefix()
		if ok2 {
			return r2, true
		}
	}

	return result, ok
}

// Flush commits any buffered or tentative state and returns the injectable result.
// The caller is responsible for re-injecting the triggering key (e.g. space) after
// applying the returned result.
func (e *Engine) Flush() (TransliterationResult, bool) {
	if e.prefixBuf != "" {
		entry, matchLen, _ := e.trie.LookupWithBuffer(e.prefixBuf)
		matched := ""
		if matchLen > 0 {
			matched = e.prefixBuf[:matchLen]
		}
		e.prefixBuf = ""

		if entry != nil {
			token := Token{Type: entry.Type, Sinhala: entry.Output, Raw: matched}
			// feedToken computes the full output (including any previously committed
			// consonants) relative to sm.prevOutput.  We treat the result as the
			// definitive committed output regardless of the Committed flag.
			result, _ := e.sm.feedToken(token)
			e.sm.reset()
			if result.Output != "" || result.DeleteCount > 0 {
				return TransliterationResult{
					Output:      result.Output,
					DeleteCount: result.DeleteCount,
					Committed:   true,
				}, true
			}
			return TransliterationResult{}, false
		}
		// No trie match for the pending chars — fall through to commit SM only.
	}

	result := e.sm.commit()
	e.sm.reset()
	if result.Output != "" || result.DeleteCount > 0 {
		return result, true
	}
	return TransliterationResult{}, false
}

// HandleBackspace handles a VK_BACK press intercepted by the hook.
//
// Returns (deleteCount, passthrough):
//   - deleteCount: number of VK_BACK events the injector should fire
//   - passthrough: if true the original backspace should reach the application;
//     if false the hook should swallow it and inject deleteCount backspaces instead
func (e *Engine) HandleBackspace() (deleteCount int, passthrough bool) {
	if e.prefixBuf != "" {
		// The buffered chars were never injected — discard the last one and
		// swallow the backspace so no visible character is accidentally deleted.
		runes := []rune(e.prefixBuf)
		e.prefixBuf = string(runes[:len(runes)-1])
		return 0, false
	}

	if e.sm.prevOutput != "" {
		// Erase the tentatively injected Sinhala, then reset SM.
		dc := len([]rune(e.sm.prevOutput))
		e.sm.reset()
		return dc, false
	}

	// Nothing pending — let the original backspace reach the application.
	return 0, true
}

// HasPendingState reports whether the engine has buffered or tentative content.
// The hook uses this to decide whether to intercept VK_BACK.
func (e *Engine) HasPendingState() bool {
	return e.prefixBuf != "" || e.sm.prevOutput != ""
}

// Reset clears all buffered and tentative state.
func (e *Engine) Reset() {
	e.prefixBuf = ""
	e.sm.reset()
}
