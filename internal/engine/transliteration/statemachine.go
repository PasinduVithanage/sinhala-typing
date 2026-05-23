package transliteration

type State int

const (
	StateIdle          State = iota
	StateConsonant
	StateConsonantCluster
	StateVowelSign
	StateComplete
)

type Event int

const (
	EventConsonant  Event = iota
	EventVowel
	EventVowelSign
	EventHal
	EventKombuwa
	EventSpace
	EventBackspace
	EventOther
)

type Token struct {
	Type    Event
	Sinhala rune
	Raw     string
}

type TransliterationResult struct {
	Output      string
	DeleteCount int
	Committed   bool
}

type StateMachine struct {
	state      State
	buffer     []Token
	mapping    MappingTable
	config     Config
	prevOutput string // Unicode chars last tentatively injected (tracks what to delete on next inject)
}

func NewStateMachine(mapping MappingTable, config Config) *StateMachine {
	return &StateMachine{
		state:   StateIdle,
		buffer:  make([]Token, 0, 8),
		mapping: mapping,
		config:  config,
	}
}

// feedToken appends a pre-identified token directly to the buffer and runs the state machine.
// Used by Engine.Feed which handles tokenization separately via the trie.
func (sm *StateMachine) feedToken(t Token) (TransliterationResult, bool) {
	sm.buffer = append(sm.buffer, t)
	return sm.transition(t)
}

// ProcessKey looks up rawInput and feeds the resulting token — kept for compatibility.
func (sm *StateMachine) ProcessKey(rawInput string) (TransliterationResult, bool) {
	token, ok := sm.mapping.Lookup(rawInput)
	if !ok {
		result := sm.commit()
		sm.reset()
		return result, true
	}
	return sm.feedToken(token)
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
		output := string(t.Sinhala)
		deleteCount := len([]rune(sm.prevOutput))
		sm.prevOutput = output
		return TransliterationResult{
			Output:      output,
			DeleteCount: deleteCount,
			Committed:   false,
		}, true

	case EventVowel:
		sm.state = StateComplete
		output := string(t.Sinhala)
		deleteCount := len([]rune(sm.prevOutput))
		sm.prevOutput = ""
		return TransliterationResult{
			Output:      output,
			DeleteCount: deleteCount,
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
		deleteCount := len([]rune(sm.prevOutput))
		sm.prevOutput = ""
		return TransliterationResult{
			Output:      composed,
			DeleteCount: deleteCount,
			Committed:   true,
		}, true

	case EventVowel:
		// In consonant context, check VowelSignMap to produce the correct vowel sign
		if len(t.Raw) > 0 {
			if sign, ok := VowelSignMap[rune(t.Raw[0])]; ok {
				sm.state = StateVowelSign
				composed := sm.composeVowelSign(sm.lastConsonant(), sign)
				deleteCount := len([]rune(sm.prevOutput))
				sm.prevOutput = ""
				return TransliterationResult{
					Output:      composed,
					DeleteCount: deleteCount,
					Committed:   true,
				}, true
			}
		}
		// No vowel sign form: commit the consonant, then output the independent vowel
		prev := sm.commitBuffer()
		deleteCount := len([]rune(sm.prevOutput))
		sm.reset()
		sm.prevOutput = ""
		return TransliterationResult{
			Output:      prev + string(t.Sinhala),
			DeleteCount: deleteCount,
			Committed:   true,
		}, true

	case EventKombuwa:
		sm.state = StateVowelSign
		output := string(sm.lastConsonant()) + string(t.Sinhala)
		deleteCount := len([]rune(sm.prevOutput))
		sm.prevOutput = output
		return TransliterationResult{
			Output:      output,
			DeleteCount: deleteCount,
			Committed:   false,
		}, true

	case EventHal:
		sm.state = StateConsonantCluster
		return TransliterationResult{}, false

	case EventConsonant:
		// Two consonants in a row — commit the previous syllable, start a new one.
		// feedToken already appended the new consonant so buffer[:len-1] is the old
		// syllable and buffer[len-1] is the newly arrived consonant.
		var prevRunes []rune
		for _, tok := range sm.buffer[:len(sm.buffer)-1] {
			prevRunes = append(prevRunes, tok.Sinhala)
		}
		prev := string(prevRunes)
		deleteCount := len([]rune(sm.prevOutput))
		sm.reset()
		sm.buffer = append(sm.buffer, t)
		sm.state = StateConsonant
		output := prev + string(t.Sinhala)
		sm.prevOutput = string(t.Sinhala) // only the new consonant is tentative
		return TransliterationResult{
			Output:      output,
			DeleteCount: deleteCount,
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
	deleteCount := len([]rune(sm.prevOutput))
	sm.state = StateConsonant
	sm.prevOutput = cluster
	return TransliterationResult{
		Output:      cluster,
		DeleteCount: deleteCount,
		Committed:   false,
	}, true
}

func (sm *StateMachine) fromVowelSign(t Token) (TransliterationResult, bool) {
	// Additional matra after kombuwa extends the vowel (e.g., ෙ + ා = ො)
	if t.Type == EventVowelSign && sm.hasKombuwa() {
		composed := string(sm.lastConsonant()) + string('ෙ') + string(t.Sinhala)
		deleteCount := len([]rune(sm.prevOutput))
		sm.state = StateComplete
		sm.prevOutput = ""
		return TransliterationResult{
			Output:      composed,
			DeleteCount: deleteCount,
			Committed:   true,
		}, true
	}
	// EventVowel 'a' after kombuwa state produces ො (kombuwa + ා)
	if t.Type == EventVowel && sm.hasKombuwa() {
		if len(t.Raw) > 0 {
			if sign, ok := VowelSignMap[rune(t.Raw[0])]; ok {
				composed := string(sm.lastConsonant()) + string('ෙ') + string(sign)
				deleteCount := len([]rune(sm.prevOutput))
				sm.state = StateComplete
				sm.prevOutput = ""
				return TransliterationResult{
					Output:      composed,
					DeleteCount: deleteCount,
					Committed:   true,
				}, true
			}
		}
	}
	result := sm.commit()
	sm.reset()
	return result, true
}

const (
	AlLakuna = '්'
	ZWJ      = '‍'
	ZWNJ     = '‌'
	Ra       = 'ර'
)

func (sm *StateMachine) buildCluster(next rune) string {
	prev := sm.lastConsonant()
	if sm.config.UseZWJClusters {
		if prev == Ra || next == Ra {
			return string(prev) + string(AlLakuna) + string(ZWJ) + string(next)
		}
		// Yansaya (ය) also uses ZWJ
		if next == 'ය' {
			return string(prev) + string(AlLakuna) + string(ZWJ) + string(next)
		}
	}
	return string(prev) + string(AlLakuna) + string(next)
}

func (sm *StateMachine) composeVowelSign(consonant, sign rune) string {
	if sm.hasKombuwa() {
		return string(consonant) + string('ෙ') + string(sign)
	}
	return string(consonant) + string(sign)
}

// lastConsonant returns the most recent consonant in the buffer, excluding
// the token at buffer[len-1] which was just appended by feedToken before
// transition() was called.
func (sm *StateMachine) lastConsonant() rune {
	for i := len(sm.buffer) - 2; i >= 0; i-- {
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
	deleteCount := len([]rune(sm.prevOutput))
	if sm.prevOutput != "" {
		// Confirm what was tentatively injected (e.g. a consonant or cluster).
		// Using prevOutput as the output preserves any ZWJ/ZWNJ that were part
		// of the tentative inject but are absent from the raw token Sinhala values.
		output := sm.prevOutput
		sm.prevOutput = ""
		return TransliterationResult{Output: output, DeleteCount: deleteCount, Committed: true}
	}
	// Nothing tentatively injected yet — output the raw buffer content.
	output := sm.commitBuffer()
	return TransliterationResult{Output: output, DeleteCount: 0, Committed: true}
}

func (sm *StateMachine) reset() {
	sm.state = StateIdle
	sm.buffer = sm.buffer[:0]
	sm.prevOutput = ""
}
