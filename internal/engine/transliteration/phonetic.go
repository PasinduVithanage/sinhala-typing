package transliteration

type MappingEntry struct {
	Output rune
	Type   Event
	MaxLen int
}

// PhoneticMap maps English keystrokes to Sinhala Unicode codepoints.
var PhoneticMap = map[string]MappingEntry{
	// Independent vowels
	"a":  {0x0D85, EventVowel, 1},
	"aa": {0x0D86, EventVowel, 2},
	"A":  {0x0D86, EventVowel, 1},
	"ae": {0x0D87, EventVowel, 2},
	"i":  {0x0D89, EventVowel, 1},
	"ii": {0x0D8A, EventVowel, 2},
	"u":  {0x0D8B, EventVowel, 1},
	"uu": {0x0D8C, EventVowel, 2},
	"e":  {0x0D91, EventVowel, 1},
	"ee": {0x0D92, EventVowel, 2},
	"o":  {0x0D94, EventVowel, 1},
	"oo": {0x0D95, EventVowel, 2},
	"au": {0x0D96, EventVowel, 2},

	// Consonants
	"k":  {0x0D9A, EventConsonant, 1},
	"kh": {0x0D9B, EventConsonant, 2},
	"g":  {0x0D9C, EventConsonant, 1},
	"gh": {0x0D9D, EventConsonant, 2},
	"ng": {0x0D9E, EventConsonant, 2},
	"c":  {0x0DA0, EventConsonant, 1},
	"ch": {0x0DA0, EventConsonant, 2},
	"j":  {0x0DA2, EventConsonant, 1},
	"jh": {0x0DA3, EventConsonant, 2},
	"ny": {0x0DA4, EventConsonant, 2},
	"T":  {0x0DA7, EventConsonant, 1},
	"Th": {0x0DA8, EventConsonant, 2},
	"D":  {0x0DA9, EventConsonant, 1},
	"Dh": {0x0DAA, EventConsonant, 2},
	"N":  {0x0DAB, EventConsonant, 1},
	"t":  {0x0DAD, EventConsonant, 1},
	"th": {0x0DAE, EventConsonant, 2},
	"d":  {0x0DAF, EventConsonant, 1},
	"dh": {0x0DB0, EventConsonant, 2},
	"n":  {0x0DB1, EventConsonant, 1},
	"p":  {0x0DB4, EventConsonant, 1},
	"ph": {0x0DB5, EventConsonant, 2},
	"b":  {0x0DB6, EventConsonant, 1},
	"bh": {0x0DB7, EventConsonant, 2},
	"m":  {0x0DB8, EventConsonant, 1},
	"mb": {0x0DB9, EventConsonant, 2},
	"y":  {0x0DBA, EventConsonant, 1},
	"r":  {0x0DBB, EventConsonant, 1},
	"l":  {0x0DBD, EventConsonant, 1},
	"v":  {0x0DC0, EventConsonant, 1},
	"w":  {0x0DC0, EventConsonant, 1},
	"sh": {0x0DC1, EventConsonant, 2},
	"Sh": {0x0DC2, EventConsonant, 2},
	"s":  {0x0DC3, EventConsonant, 1},
	"h":  {0x0DC4, EventConsonant, 1},
	"L":  {0x0DC5, EventConsonant, 1},
	"f":  {0x0DC6, EventConsonant, 1},

	// Kombuwa (pre-base vowel sign) — handled specially in state machine
	"E": {0x0DD9, EventKombuwa, 1},

	// Al-lakuna triggers
	"H": {AlLakuna, EventHal, 1},
	"^": {AlLakuna, EventHal, 1},
}

// VowelSignMap maps raw keys (when in StateConsonant) to vowel sign codepoints.
var VowelSignMap = map[rune]rune{
	'a': 0x0DCF, // ා
	'i': 0x0DD2, // ි
	'I': 0x0DD3, // ී
	'u': 0x0DD4, // ු
	'U': 0x0DD6, // ූ
	'e': 0x0DD9, // ෙ (kombuwa — see EventKombuwa)
	'E': 0x0DDA, // ේ
	'o': 0x0DDC, // ො (kombuwa + ා — produced as sequence)
}
