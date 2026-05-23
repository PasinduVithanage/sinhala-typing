# Sinhala Assistant — How to Use

Sinhala Assistant lets you type Sinhala Unicode in **any Windows application** — Notepad, Word, Chrome, VS Code, Photoshop, and more — using your regular English keyboard. No Sinhala keyboard layout to install, no IME to configure.

---

## Installation

1. Go to the **[Releases page](https://github.com/PasinduVithanage/sinhala-typing/releases)**.
2. Under the latest release, download **`sinhala-assistant-amd64-installer.exe`**.
3. Double-click the downloaded file and follow the short wizard.
   - Default install path: `C:\Program Files\PasinduVithanage\Sinhala Assistant`
   - No admin rights are needed if you change the path to a folder you own.
4. When the wizard finishes, a shortcut appears on your **Desktop** and in the **Start Menu**.
5. Launch Sinhala Assistant — a small icon appears in the **system tray** (bottom-right of the taskbar).

---

## First Launch

After launch you will see a small icon in the **system tray** (notification area, bottom-right):

| Icon colour | Meaning |
|---|---|
| Green circle | Sinhala typing is **active** |
| Gray circle | Sinhala typing is **inactive** |
| Red circle | An error occurred |

The engine starts **inactive** by default. Enable it before you start typing.

---

## Enabling / Disabling Sinhala Typing

**Three ways to toggle:**

1. **Global hotkey** — press `Ctrl + Alt + S` from anywhere. The tray icon turns green/gray instantly.
2. **Tray menu** — right-click the tray icon → *Enable Sinhala Typing* / *Disable Sinhala Typing*.
3. **Settings panel** — open the app window, go to Settings → "Start engine on launch" saves the auto-start preference.

---

## Typing Sinhala — Phonetic Mapping

When the engine is active, every **A–Z key** you press is intercepted and converted to Sinhala using a **phonetic transliteration** scheme. The conversion happens in-place: Latin characters are replaced with the correct Sinhala Unicode glyphs automatically.

### Consonants

| Type | Sinhala | Type | Sinhala |
|---|---|---|---|
| `k` | ක | `g` | ග |
| `c` / `ch` | ච | `j` | ජ |
| `t` | ට | `d` | ඩ |
| `T` | ත | `D` | ද |
| `n` | න | `N` | ණ |
| `p` | ප | `b` | බ |
| `m` | ම | `y` | ය |
| `r` | ර | `l` | ල |
| `v` / `w` | ව | `s` | ස |
| `S` | ශ | `h` | හ |
| `f` | ෆ | `L` | ළ |

### Vowels (standalone)

| Type | Sinhala |
|---|---|
| `a` | අ |
| `aa` / `A` | ආ |
| `i` | ඉ |
| `ii` / `I` | ඊ |
| `u` | උ |
| `uu` / `U` | ඌ |
| `e` | එ |
| `ee` / `E` | ඒ |
| `o` | ඔ |
| `oo` / `O` | ඕ |

### Vowel signs (after a consonant)

Type the consonant, then the vowel modifier:

| Suffix | Sign | Example |
|---|---|---|
| `a` (default) | ් (hal) | `k` → ක් |
| `aa` | ා | `ka` → කා |
| `i` | ි | `ki` → කි |
| `ii` | ී | `kii` → කී |
| `u` | ු | `ku` → කු |
| `uu` | ූ | `kuu` → කූ |
| `e` | ෙ | `ke` → කෙ |
| `ee` | ේ | `kee` → කේ |
| `o` | ො | `ko` → කො |
| `oo` | ෝ | `koo` → කෝ |
| `au` | ෞ | `kau` → කෞ |

### Special cluster forms (ZWJ)

| Type | Sinhala |
|---|---|
| `kra` | ක්‍ර (rakar) |
| `kya` | ක්‍ය (yansaya) |

ZWJ clusters can be disabled in Settings if your target application does not render them correctly.

### Space / punctuation

Pressing **Space** flushes any pending prefix and passes the space through normally. All non-letter keys (Enter, Tab, arrow keys, punctuation) are passed through untouched.

### Backspace

Backspace is context-aware: while there is pending Sinhala state the engine handles deletion correctly (removing the right number of Unicode code points). Once the buffer is clear, Backspace behaves normally.

---

## Application Compatibility

Sinhala Assistant automatically detects which application is in focus and chooses the best injection method:

| Mode | Applications | How it works |
|---|---|---|
| **Keyboard** (default) | Notepad, Word, Chrome, Firefox, VS Code, … | SendInput Unicode injection |
| **Clipboard** | Photoshop, Illustrator, Java apps | Pastes via clipboard (Ctrl+V) |
| **Blocked** | System dialogs, UAC prompts | Typing is passed through unchanged |

You do not need to configure anything — detection is automatic.

---

## Fix Clipboard

If you copy Sinhala text from an older source that uses the Wijesekera legacy encoding, Sinhala Assistant can repair it:

1. Copy the garbled text.
2. Right-click the tray icon → **Fix Clipboard**.
3. Paste — the text is now correct Unicode.

You can also enable **Auto-fix Sinhala in clipboard** in Settings to repair automatically whenever you copy.

---

## Settings

Open the app window (double-click the tray icon or right-click → *Settings*).

| Setting | Description |
|---|---|
| Start engine on launch | Activates Sinhala typing automatically when the app starts |
| Auto-fix Sinhala in clipboard | Repairs legacy-encoded Sinhala whenever you copy text |
| Use ZWJ for cluster forms | Enables rakar/yansaya ZWJ ligatures (recommended: on) |
| Toggle hotkey | Key combo to enable/disable typing (default: `Ctrl+Alt+S`) |
| Start with Windows | Adds the app to the Windows startup registry so it launches at login |

Click **Save Settings** to apply.

---

## Start with Windows

Enable **Start with Windows** in Settings to have Sinhala Assistant launch automatically every time you log into Windows. This writes a value to `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run` — no admin rights required, and it is removed cleanly by the uninstaller.

---

## Uninstalling

Open **Windows Settings → Apps → Installed apps**, search for **Sinhala Assistant**, and click **Uninstall**. The uninstaller removes the executable, Start Menu and Desktop shortcuts, app data (`%APPDATA%\sinhala-assistant\`), and the startup registry entry.

---

## Building from Source (developers)

Requirements: **Go 1.25+**, **Wails v2 CLI**, **Node.js 18+**, **NSIS 3+** (installer only).

```powershell
# Install Wails CLI (once)
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# Clone the repo
git clone git@github.com:PasinduVithanage/sinhala-typing.git
cd sinhala-typing

# Portable EXE
wails build
# → build\bin\sinhala-assistant.exe

# NSIS installer  (requires NSIS on PATH)
wails build --nsis
# → build\bin\sinhala-assistant-amd64-installer.exe
```

---

## Troubleshooting

| Problem | Fix |
|---|---|
| Tray icon does not appear | Check Task Manager — the process may already be running |
| Typing produces Latin characters | Make sure the tray icon is green (engine active) |
| Text is not injected in a specific app | The app may be in "blocked" mode; check if it runs elevated (UAC) |
| Installer requires admin rights | The installer installs to Program Files — run as administrator |
| `Ctrl+Alt+S` conflicts with another app | Change the hotkey in Settings to a different combo |
| ZWJ clusters render as boxes | Disable "Use ZWJ for cluster forms" in Settings |

---

## Keyboard Shortcut Reference

| Action | Shortcut |
|---|---|
| Toggle Sinhala typing | `Ctrl + Alt + S` (configurable) |
| Flush pending buffer | `Space` |
| Context-aware backspace | `Backspace` |
