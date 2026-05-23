package normalization

import "strings"

// WijeskeeraToUnicode maps Wijesekera font ASCII positions to Sinhala Unicode.
var WijeskeeraToUnicode = map[rune]rune{
	'q': 0x0DCA, // ්
	'w': 0x0DD9, // ෙ
	'e': 0x0DD0, // ැ
	'r': 0x0DBB, // ර
	't': 0x0DAD, // ත
	'y': 0x0DBA, // ය
	'u': 0x0DD4, // ු
	'i': 0x0DD2, // ි
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
	'I': 0x0DD3, // ී
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
	b.Grow(len(s) * 3)
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
