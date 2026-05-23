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

// RepairString detects legacy encoding first, then applies Unicode repair rules.
func RepairString(s string) string {
	if IsLikelyLegacyEncoded(s) {
		s = ConvertLegacyWijesekera(s)
	}
	if len(Analyze(s)) == 0 {
		return s
	}
	return Repair(s)
}
