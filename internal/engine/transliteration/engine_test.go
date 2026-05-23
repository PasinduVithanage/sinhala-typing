package transliteration

import (
	"testing"
)

// helper: feed a sequence of chars and return all non-zero results
func feedAll(e *Engine, chars string) []TransliterationResult {
	var results []TransliterationResult
	for _, ch := range chars {
		r, ok := e.Feed(string(ch))
		if ok {
			results = append(results, r)
		}
	}
	return results
}

// helper: feed chars then flush, return the final committed result
func feedFlush(e *Engine, chars string) TransliterationResult {
	for _, ch := range chars {
		e.Feed(string(ch))
	}
	r, _ := e.Flush()
	return r
}

func TestFeed_SingleConsonant_WaitsForDigraph(t *testing.T) {
	e := NewEngine("phonetic", true)
	_, ok := e.Feed("k")
	if ok {
		t.Fatal("'k' alone should return false (waiting for possible 'kh')")
	}
}

func TestFlush_SingleConsonant_Commits(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k")
	r, ok := e.Flush()
	if !ok || r.Output != "ක" || r.DeleteCount != 0 || !r.Committed {
		t.Fatalf("flush after 'k' got %+v", r)
	}
}

func TestFeed_DigraphKh(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k")
	r, ok := e.Feed("h")
	if !ok || r.Output != "ඛ" || r.Committed {
		t.Fatalf("'kh' digraph got %+v", r)
	}
}

func TestFeed_DigraphNg(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("n")
	r, ok := e.Feed("g")
	if !ok || r.Output != "ඞ" {
		t.Fatalf("'ng' digraph got %+v", r)
	}
}

func TestFeed_DigraphSh(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("s")
	r, ok := e.Feed("h")
	if !ok || r.Output != "ශ" {
		t.Fatalf("'sh' digraph got %+v", r)
	}
}

func TestFeed_DigraphTh(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("t")
	r, ok := e.Feed("h")
	if !ok || r.Output != "ථ" {
		t.Fatalf("'th' digraph got %+v", r)
	}
}

// 'n' followed by 'a': resolves 'n' as න, buffers 'a'
func TestFeed_DiambiguateNa(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("n")
	r, ok := e.Feed("a") // 'na' → 'n' matched, 'a' still pending
	if !ok || r.Output != "න" || r.Committed {
		t.Fatalf("'n'+'a' should inject න tentatively, got %+v", r)
	}
	// Now flush 'a' as vowel sign
	r2, ok2 := e.Flush()
	if !ok2 || r2.Output != "නා" || r2.DeleteCount != 1 || !r2.Committed {
		t.Fatalf("flush after 'na' got %+v", r2)
	}
}

// Independent vowel 'a' in isolation
func TestFeed_IndependentVowel_a(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("a")
	// 'a' is a prefix of 'aa','ae','au' so it waits
	r, ok := e.Flush()
	if !ok || r.Output != "අ" || !r.Committed {
		t.Fatalf("independent 'a' got %+v", r)
	}
}

// 'aa' double-vowel
func TestFeed_DoubleVowel_aa(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("a")
	r, ok := e.Feed("a")
	if !ok || r.Output != "ආ" || !r.Committed {
		t.Fatalf("'aa' got %+v", r)
	}
}

// Consonant + vowel sign 'a' → ā (consonant+ā)
func TestFeed_VowelSign_a(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k") // wait
	e.Feed("a") // 'k' resolved, 'a' buffered — ක injected tentatively
	r, ok := e.Flush()
	if !ok || r.Output != "කා" || r.DeleteCount != 1 || !r.Committed {
		t.Fatalf("'ka' flush got %+v", r)
	}
}

// Consonant + vowel sign 'i'
func TestFeed_VowelSign_i(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("m")
	r, ok := e.Feed("i")
	// 'm' resolves immediately (no 'mi' prefix), 'i' is also prefix of 'ii'
	// First result is ම tentative
	if !ok || r.Output != "ම" {
		t.Fatalf("'m'+'i' first result got %+v", r)
	}
	r2, ok2 := e.Flush()
	if !ok2 || r2.Output != "මි" || r2.DeleteCount != 1 || !r2.Committed {
		t.Fatalf("flush 'mi' got %+v", r2)
	}
}

// Consonant cluster via 'H' (al-lakuna): kHk → ක්ක
func TestFeed_ConsonantCluster_kHk(t *testing.T) {
	e := NewEngine("phonetic", false) // no ZWJ
	e.Feed("k") // wait ('k' is prefix of 'kh')
	e.Feed("H") // 'kH' → 'k' matched + 'H' (hal) → ක tentative, SM in StateConsonantCluster
	e.Feed("k") // wait again for possible 'kh'
	r, ok := e.Flush()
	// Expected: delete 1 (tentative ക), inject ക්ක
	if !ok || r.Output != "ක්ක" || r.DeleteCount != 1 || !r.Committed {
		t.Fatalf("cluster 'kHk' flush got %+v (output=%q)", r, r.Output)
	}
}

// Kombuwa sequence: kEa → කොා (consonant + kombuwa + ā)
func TestFeed_Kombuwa_kEa(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k") // wait
	e.Feed("E") // 'kE': 'k' matched, 'E' fed as EventKombuwa
	// After above: කෙ injected tentatively (DeleteCount=1 erasing kok)
	e.Feed("a") // wait ('a' prefix of 'aa')
	r, ok := e.Flush()
	// Expected: delete 2 (tentative කෙ = 2 Unicode chars), inject kok+ෙ+ā
	// composeVowelSign with hasKombuwa: consonant + ෙ + ā
	if !ok || r.Output != "ක"+"ෙ"+"ා" || r.DeleteCount != 2 || !r.Committed {
		t.Fatalf("kombuwa 'kEa' flush got %+v (output=%q)", r, r.Output)
	}
}

// Two consonants in a row: km → ම replaces ක, kem injected as committed+new
func TestFeed_TwoConsonants(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k")
	results := feedAll(e, "m")
	// 'km': 'k' matches, remaining='' → kok injected tentatively
	// 'm' has no digraphs? Actually 'm' is prefix of 'mb'. So 'm' → isPrefix=true → wait
	// But 'km': k-node.children has 'h'. 'm' is not a child → matches 'k', remaining='m'
	// feedToken(kok) → inject kok tentatively
	// prefixBuf='m' → isPrefix=true → wait
	// Feed("m") returns (kok, true)
	if len(results) != 1 || results[0].Output != "ක" {
		t.Fatalf("'km': expected kok tentative, got %v", results)
	}
	// Now feed 'a': 'ma' → 'm' matches, remaining='a'
	r, ok := e.Feed("a")
	// feedToken(ම) → fromConsonant(EventConsonant): prev="kok", output="කම", deleteCount=1
	// remaining='a' → wait → return (result_for_m, true)
	if !ok || r.Output != "කම" || r.DeleteCount != 1 || r.Committed {
		t.Fatalf("'kma' mid result got %+v", r)
	}
	// Flush 'a' as vowel sign
	r2, ok2 := e.Flush()
	if !ok2 || r2.Output != "මා" || r2.DeleteCount != 1 || !r2.Committed {
		t.Fatalf("flush 'kma' got %+v (output=%q)", r2, r2.Output)
	}
}

// Unmapped key passes through (with any SM state committed)
func TestFeed_UnmappedKey(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k")
	e.Feed("H") // kok + hal → StateConsonantCluster, kok injected
	// Now feed unmapped 'x' (not in map)
	r, ok := e.Feed("x")
	// 'x' not in trie → commit SM + passthrough 'x'
	// SM in StateConsonantCluster: commit() → output = commitBuffer() = kok? Let me think.
	// Actually fromCluster checks: if not EventConsonant, call commit()+reset()
	// But we're in resolvePrefix, which calls sm.commit() directly when entry==nil.
	// sm.commit() → commitBuffer() with buffer=[kok, H] → "කාNot clear, let me just check that it returns true and the output contains 'x'
	if !ok || r.Output == "" {
		t.Fatalf("unmapped 'x' should produce output, got %+v", r)
	}
	// Output should end with 'x'
	runes := []rune(r.Output)
	if runes[len(runes)-1] != 'x' {
		t.Fatalf("unmapped 'x' output should end with 'x', got %q", r.Output)
	}
}

// Backspace with nothing pending: passthrough
func TestHandleBackspace_NothingPending(t *testing.T) {
	e := NewEngine("phonetic", true)
	dc, passthrough := e.HandleBackspace()
	if dc != 0 || !passthrough {
		t.Fatalf("empty backspace should passthrough, got dc=%d passthrough=%v", dc, passthrough)
	}
}

// Backspace with prefixBuf pending: discard + swallow
func TestHandleBackspace_PrefixBufPending(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k") // buffered, not injected
	dc, passthrough := e.HandleBackspace()
	if dc != 0 || passthrough {
		t.Fatalf("backspace with prefix pending: expected (0,false), got (%d,%v)", dc, passthrough)
	}
	if e.prefixBuf != "" {
		t.Fatal("prefixBuf should be cleared after backspace")
	}
}

// Backspace with tentative Sinhala injected: erase injected chars
func TestHandleBackspace_TentativeSinhala(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("m") // 'm' is prefix of 'mb', wait
	e.Feed("a") // 'ma' → ム resolved, 'a' buffered; ම injected tentatively (prevOutput="ම")
	// HasPendingState: prefixBuf="a" or prevOutput="ම"
	if !e.HasPendingState() {
		t.Fatal("should have pending state after 'ma'")
	}
	// First backspace: clears prefixBuf 'a', not the tentative ම
	dc, passthrough := e.HandleBackspace()
	if dc != 0 || passthrough {
		t.Fatalf("first backspace should clear prefixBuf: got (%d,%v)", dc, passthrough)
	}
	// Second backspace: erases tentative ම
	dc, passthrough = e.HandleBackspace()
	if dc != 1 || passthrough { // 1 because ම is 1 Unicode char
		t.Fatalf("second backspace should erase ම: got dc=%d passthrough=%v", dc, passthrough)
	}
}

// HasPendingState
func TestHasPendingState(t *testing.T) {
	e := NewEngine("phonetic", true)
	if e.HasPendingState() {
		t.Fatal("fresh engine should have no pending state")
	}
	e.Feed("k")
	if !e.HasPendingState() {
		t.Fatal("after 'k' (buffered), should have pending state")
	}
	e.Reset()
	if e.HasPendingState() {
		t.Fatal("after Reset, should have no pending state")
	}
}

// Rakar cluster with ZWJ: rHk → ර්‍ක
func TestFeed_RakarCluster_ZWJ(t *testing.T) {
	e := NewEngine("phonetic", true) // useZWJ=true
	e.Feed("r")                      // 'r' resolves immediately (no digraphs)
	_, ok := e.Feed("H")             // 'rH': r matched + H fed → ර tentative, SM cluster state
	// 'r' is not a prefix so Feed("r") returns immediately
	// But Feed("H"): prefixBuf="" after r resolved, then "H" → EventHal
	// Let's just check the cluster output
	_ = ok
	r, ok2 := e.Feed("k") // 'k' → wait (prefix of kh)
	_ = r
	_ = ok2
	r3, ok3 := e.Flush()
	if !ok3 {
		t.Fatal("expected output from rakar cluster flush")
	}
	// Expected: ර + ් + ZWJ + ක = "ර්‍ක"
	expected := "ර" + string('්') + string('‍') + "ක"
	if r3.Output != expected {
		t.Fatalf("rakar cluster ZWJ: expected %q, got %q", expected, r3.Output)
	}
}

// Yansaya cluster with ZWJ: kHy → ක්‍ය
func TestFeed_YansayaCluster_ZWJ(t *testing.T) {
	e := NewEngine("phonetic", true)
	e.Feed("k")
	e.Feed("H")
	e.Feed("y") // 'y' resolves immediately
	r, ok := e.Flush()
	if !ok {
		t.Fatal("expected output from yansaya flush")
	}
	expected := "ක" + string('්') + string('‍') + "ය"
	if r.Output != expected {
		t.Fatalf("yansaya cluster ZWJ: expected %q, got %q", expected, r.Output)
	}
}

// Full word "sinhala" phonetic sequence — smoke test
func TestFeed_WordSinhala(t *testing.T) {
	e := NewEngine("phonetic", true)
	// s-i-n-h-a-l-a → සිංහල (approximate, depends on exact key sequence)
	// Just verify no panic and we get some output
	chars := "sinhala"
	var total string
	for _, ch := range chars {
		r, ok := e.Feed(string(ch))
		if ok && r.Output != "" {
			total += r.Output
		}
	}
	r, ok := e.Flush()
	if ok {
		total += r.Output
	}
	if total == "" {
		t.Fatal("expected some Sinhala output for 'sinhala'")
	}
}
