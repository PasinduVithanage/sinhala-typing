package keyboard

import (
	"fmt"
	"unicode"
)

// vkToASCII converts a virtual key code + shift state to an ASCII character string.
// Returns empty string for non-printable keys.
func vkToASCII(vk uint32, shift bool) string {
	// Letter keys VK_A–VK_Z are 0x41–0x5A
	if vk >= 0x41 && vk <= 0x5A {
		ch := rune(vk)
		if !shift {
			ch = unicode.ToLower(ch)
		}
		return string(ch)
	}
	// Number row 0–9 are VK 0x30–0x39
	if vk >= 0x30 && vk <= 0x39 {
		if shift {
			shifted := "!@#$%^&*()"
			return string(shifted[vk-0x30])
		}
		return fmt.Sprintf("%c", '0'+vk-0x30)
	}
	// Common punctuation
	punctMap := map[uint32][2]string{
		0xBB: {"=", "+"},
		0xBD: {"-", "_"},
		0xBE: {".", ">"},
		0xBC: {",", "<"},
		0xBF: {"/", "?"},
		0xBA: {";", ":"},
		0xDE: {"'", "\""},
		0xDB: {"[", "{"},
		0xDD: {"]", "}"},
		0xDC: {"\\", "|"},
		0xC0: {"`", "~"},
	}
	if pair, ok := punctMap[vk]; ok {
		if shift {
			return pair[1]
		}
		return pair[0]
	}
	return ""
}

// IsModifierKey returns true for modifier virtual key codes.
func IsModifierKey(vk uint32) bool {
	return vk == VK_SHIFT || vk == VK_CTRL || vk == VK_MENU ||
		vk == 0xA0 || vk == 0xA1 || // L/R shift
		vk == 0xA2 || vk == 0xA3 || // L/R ctrl
		vk == 0xA4 || vk == 0xA5    // L/R alt
}

// KeyEventToRaw converts a KeyEvent to the raw string representation.
func KeyEventToRaw(e KeyEvent) string {
	if IsModifierKey(e.VkCode) {
		return ""
	}
	return vkToASCII(e.VkCode, e.Shift)
}
