package composition

import "golang.org/x/text/unicode/norm"

// Syllable represents a decomposed Sinhala syllable.
type Syllable struct {
	Base     rune   // consonant or independent vowel
	Cluster  []rune // additional consonants joined by virama
	Kombuwa  bool   // ෙ present
	VowelSign rune  // matra after base
}

// Composer assembles Sinhala syllable components into correct Unicode sequences.
type Composer struct{}

func New() *Composer { return &Composer{} }

// Compose produces the canonical Unicode string for a syllable.
func (c *Composer) Compose(s Syllable) string {
	var out []rune
	out = append(out, s.Base)
	for i, consonant := range s.Cluster {
		out = append(out, 0x0DCA) // al-lakuna
		// Add ZWJ for rakar/repaya/yansaya forms
		if i == len(s.Cluster)-1 && (consonant == 0x0DBB || consonant == 0x0DBA) {
			out = append(out, 0x200D)
		}
		out = append(out, consonant)
	}
	if s.Kombuwa {
		out = append(out, 0x0DD9)
	}
	if s.VowelSign != 0 {
		out = append(out, s.VowelSign)
	}
	return norm.NFC.String(string(out))
}

// Validate checks whether a Unicode sequence represents a well-formed Sinhala syllable.
func (c *Composer) Validate(s string) bool {
	runes := []rune(norm.NFC.String(s))
	if len(runes) == 0 {
		return false
	}
	for i, r := range runes {
		if r == 0x0DD9 { // kombuwa
			if i == 0 || !isSinhalaConsonant(runes[i-1]) {
				return false
			}
		}
	}
	return true
}

func isSinhalaConsonant(r rune) bool { return r >= 0x0D9A && r <= 0x0DC6 }
