# Sinhala Assistant — Project Status

> Last updated: 2026-05-23

---

## Current Stage: Phase 4 Complete · Ready to Build & Release

All four phases are implemented. Run `wails build --nsis` to produce the installer.

```
✅ Phase 1 — Core Engine              COMPLETE
✅ Phase 2 — Wails App + Tray         COMPLETE
✅ Phase 3 — Keyboard Hook            COMPLETE
✅ Phase 4 — Polish & Release         COMPLETE
```

---

## What Is Built

### Go Backend — 27 source files across 12 packages

| Package | Status | Notes |
|---|---|---|
| `internal/engine/transliteration` | ✅ Complete | Prefix-aware Engine (digraphs, Flush, HandleBackspace), FSM with prevOutput tracking, ZWJ clusters, 22 unit tests |
| `internal/engine/normalization` | ✅ Complete | Detector, repair engine, Wijesekera converter |
| `internal/engine/composition` | ✅ Complete | Syllable composer, sequence validator |
| `internal/hook/keyboard` | ✅ Complete | Selective intercept (A–Z + space), VK_BACK conditional intercept, prefix-aware Processor, InjectionGuard, per-app compat routing, `InjectViaClipboard` |
| `internal/hook/clipboard` | ✅ Complete | Windows clipboard read/write/monitor |
| `internal/window` | ✅ Complete | Foreground window detection, per-app compat modes (13 apps) |
| `internal/hotkey` | ✅ Complete | Global hotkey via `RegisterHotKey` / `WM_HOTKEY`, parses "Ctrl+Alt+S" combos, `PostThreadMessage` stop |
| `internal/tray` | ✅ Complete | System tray icon, menu, icon state |
| `internal/config` | ✅ Complete | JSON config persistence in `%APPDATA%` |
| `internal/ipc` | ✅ Complete | Wails IPC bridge; hotkey wired in Startup/Shutdown/SaveConfig; double-start guard via `engineRunning` atomic |
| `internal/logging` | ✅ Complete | Structured zap logger |
| `internal/startup` | ✅ Complete | Windows `HKCU\Run` registry read/write; non-Windows stub |

### React Frontend — TypeScript throughout, zero `@ts-ignore`

| Component | Status | Notes |
|---|---|---|
| `App.tsx` | ✅ Complete | View routing (main/settings), loads config on startup, listens for `ui:show-settings` Wails event |
| `TrayPanel` | ✅ Complete | Props-based settings navigation, `CSSProperties` typed drag region |
| `TextAnalyzer` | ✅ Complete | Uses `normalization.Detection` from generated models |
| `SettingsPanel` | ✅ Complete | Includes `mapping` from `engineStore` in `SaveConfig` |
| `engineStore` (Zustand) | ✅ Built | Active state, mapping, error |
| `settingsStore` (Zustand) | ✅ Built | All user settings |
| `useEngine` hook | ✅ Built | Wails event subscriptions |

### Phase 4 Changes (this session)

- **`internal/hook/keyboard/processor.go`** — Added `done chan struct{}` field and `Stop()` method. Changed `Run()` from a blocking `range` loop to a `select`-based loop; exits cleanly on `Stop()` or channel close. Extracted inner logic to `process()`.
- **`internal/ipc/bridge.go`** — `ToggleEngine` now calls `processor.Stop()` before clearing `engineRunning`. `startEngine` creates a fresh `Processor` (new done channel) on every start cycle. Imports `internal/startup`; calls `startup.Set()` in both `Startup` and `SaveConfig`.
- **`internal/tray/icons.go`** — New file. Generates anti-aliased 22×22 PNG tray icons at runtime using `image/png`. Green = active, gray = inactive, red = error.
- **`internal/tray/tray.go`** — Removed 1×1 placeholder `iconBytes`/`minimalPNG` functions (moved to `icons.go`).
- **`internal/startup/startup_windows.go`** — New package. Reads/writes `HKCU\…\Run\SinhalaAssistant` via `golang.org/x/sys/windows/registry`. `Set(true)` writes the exe path; `Set(false)` deletes the value.
- **`internal/startup/startup_other.go`** — Non-Windows stub.
- **`internal/config/config.go`** — Added `RunAtStartup bool` (`"run_at_startup"` JSON key).
- **`build/windows/installer/project.nsi`** — Pinned product name/version/company defines. Uninstall section now deletes the `SinhalaAssistant` Run registry value and `%APPDATA%\sinhala-assistant\`.
- **`wails.json`** — Added `info` block (companyName, productName, productVersion, copyright, comments).
- **Frontend** — `models.ts` adds `run_at_startup`. `settingsStore.ts` adds `runAtStartup`/`setRunAtStartup`. `SettingsPanel.tsx` adds "Start with Windows" toggle and includes `run_at_startup` in `SaveConfig`. `App.tsx` loads `run_at_startup` from `GetConfig`.
- **`HOW_TO_USE.md`** — New user guide: installation, phonetic mapping table, compat modes, Fix Clipboard, Settings reference, troubleshooting.

### Phase 3 Changes (previous session)

- **`internal/hotkey/hotkey.go`** — New package. `RegisterHotKey` with parsed combo string (e.g. `"Ctrl+Alt+S"` → `MOD_CONTROL | MOD_ALT` + VK `0x53`). Message loop uses `LockOSThread` + `GetMessageW`; `Stop()` wakes the loop via `PostThreadMessageW(WM_NULL)`. Thread ID stored atomically so `Stop()` can target the locked OS thread.
- **`internal/hook/keyboard/injector.go`** — Added `InjectViaClipboard(text, deleteCount)`: saves prior clipboard, sets new Sinhala text, sends Ctrl+V via `SendInput`, restores prior clipboard after 300ms. Added `injectCtrlV()` helper.
- **`internal/hook/keyboard/processor.go`** — Now calls `foregroundCompatMode()` per keystroke (`GetForegroundWindowInfo` + `GetCompatMode`). `inject()` routes to `InjectViaClipboard` for `CompatModeClipboard` apps (Photoshop, Illustrator, Java), suppresses injection for `CompatModeBlocked`, uses keyboard injection for everything else.
- **`internal/ipc/bridge.go`** — Added `hk *hotkey.Hotkey` field. `Startup` starts the hotkey goroutine. `Shutdown` stops it. `SaveConfig` re-registers the hotkey when `toggle_hotkey` changes. Added `engineRunning atomic.Bool` guard in `startEngine` to prevent double-goroutine-start if `ToggleEngine` is called concurrently; `ToggleEngine` resets the flag on stop.

---

## Test Suite

```
ok  sinhala-assistant/internal/engine/transliteration   22 tests pass
```

---

## Build Artifacts

```
build/bin/sinhala-assistant.exe   ← production binary (Windows amd64)
```

Built with: `wails build` · Wails v2.12.0 · Go 1.25 · React 18 · Tailwind CSS v4

---

## Remaining Gaps

All Phase 4 items are resolved. No known gaps remain for a v1.0 release.

To produce release artifacts:
```powershell
wails build                 # portable EXE → build/bin/sinhala-assistant.exe
wails build --nsis          # NSIS installer → build/bin/sinhala-assistant-amd64-installer.exe
```

Optional future work:
- Code-signing the EXE and installer (removes SmartScreen warning)
- Auto-update mechanism
- Additional mapping profiles (Wijesekera layout as an alternative to phonetic)

---

## Architecture Reference

Full architecture document: [`docs/architecture.md`](docs/architecture.md)

### Key Design Rules
- **`internal/engine/` is pure Go** — zero Windows API calls, unit-testable on any OS.
- **`internal/hook/` and `internal/hotkey/` require `//go:build windows`** — use `golang.org/x/sys/windows`.
- **Digraph buffering** — `Engine.prefixBuf` holds raw chars until the trie confirms a definitive match. Nothing is injected while waiting; a non-letter key (space) triggers `Flush()`.
- **`prevOutput` tracking** — `StateMachine.prevOutput` stores the Unicode string last tentatively injected. `DeleteCount` is always `len([]rune(prevOutput))`, never raw ASCII byte count.
- **Kombuwa ordering** — engine stores kombuwa AFTER the base consonant (U+0DD9 follows consonant).
- **ZWJ clusters** — rakar `ක්‍ර`, yansaya `ක්‍ය`, repaya `ར්‍ක` require ZWJ (U+200D). `commit()` uses `prevOutput` (not `commitBuffer()`) to preserve injected ZWJ.
- **Injection guard** — `InjectionGuard` limits to 50 `SendInput` calls/second.
- **Compat mode routing** — `foregroundCompatMode()` is called per keystroke; clipboard apps get `InjectViaClipboard`, blocked apps get nothing, all others get `ReplaceWithSinhala`.
- **Hotkey thread** — `hotkey.Run()` calls `runtime.LockOSThread()` so `RegisterHotKey` and its `GetMessageW` loop share one OS thread. `Stop()` posts `WM_NULL` to wake the loop.

---

## Stack

| Layer | Technology |
|---|---|
| Runtime | Go 1.25, Wails v2.12.0 |
| Windows API | `golang.org/x/sys/windows` |
| Unicode | `golang.org/x/text/unicode/norm` |
| System tray | `github.com/getlantern/systray` |
| Config | `encoding/json` (stdlib) |
| Logging | `go.uber.org/zap` |
| Frontend | React 18 + TypeScript + Vite |
| Styling | Tailwind CSS v4 (`@tailwindcss/postcss`) |
| State | Zustand |
| IPC | Wails auto-generated bindings |
