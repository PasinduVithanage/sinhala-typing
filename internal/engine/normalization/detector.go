package normalization

import "golang.org/x/text/unicode/norm"

// Issue classifies a Sinhala Unicode rendering problem.
type Issue int

const (
	IssueNone              Issue = iota
	IssueKombuwaBeforeBase
	IssueVowelSignBeforeBase
	IssueDuplicateVirama
	IssueOrphanedVirama
	IssueWrongClusterOrder
	IssueLegacyEncoding
)

// Detection describes one located issue in the input string.
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
					Context: string(runes[clamp(i-1, 0, len(runes)):clamp(i+2, 0, len(runes))]),
				})
			}
		}

		// Any vowel sign must follow a consonant
		if isSinhalaVowelSign(r) && r != 0x0DD9 {
			if i == 0 || (!isSinhalaConsonant(runes[i-1]) && runes[i-1] != 0x0DCA) {
				issues = append(issues, Detection{
					StartIndex: i, EndIndex: i + 1,
					Issue:   IssueVowelSignBeforeBase,
					Context: string(runes[clamp(i-1, 0, len(runes)):clamp(i+2, 0, len(runes))]),
				})
			}
		}

		// Double virama
		if r == 0x0DCA && i+1 < len(runes) && runes[i+1] == 0x0DCA {
			issues = append(issues, Detection{StartIndex: i, EndIndex: i + 2, Issue: IssueDuplicateVirama})
		}

		// Orphaned virama (not preceded by a consonant)
		if r == 0x0DCA && (i == 0 || !isSinhalaConsonant(runes[i-1])) {
			issues = append(issues, Detection{StartIndex: i, EndIndex: i + 1, Issue: IssueOrphanedVirama})
		}
	}

	return issues
}

func isSinhalaConsonant(r rune) bool { return r >= 0x0D9A && r <= 0x0DC6 }
func isSinhalaVowelSign(r rune) bool  { return r >= 0x0DCF && r <= 0x0DDF }

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
