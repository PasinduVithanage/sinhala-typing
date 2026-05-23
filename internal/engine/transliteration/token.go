package transliteration

// TrieNode is a node in the prefix lookup trie.
type TrieNode struct {
	children map[rune]*TrieNode
	entry    *MappingEntry
}

// Trie provides O(k) prefix lookup for mapping keys.
type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(key string, entry MappingEntry) {
	node := t.root
	for _, ch := range key {
		child, ok := node.children[ch]
		if !ok {
			child = &TrieNode{children: make(map[rune]*TrieNode)}
			node.children[ch] = child
		}
		node = child
	}
	node.entry = &entry
}

// LookupWithBuffer finds the longest matching prefix in buf.
// isPrefix=true means buf could be the start of a longer valid key.
func (t *Trie) LookupWithBuffer(buf string) (entry *MappingEntry, matchLen int, isPrefix bool) {
	node := t.root
	lastMatch := (*MappingEntry)(nil)
	lastMatchLen := 0

	for i, ch := range buf {
		child, ok := node.children[ch]
		if !ok {
			return lastMatch, lastMatchLen, false
		}
		node = child
		if node.entry != nil {
			lastMatch = node.entry
			lastMatchLen = i + 1
		}
	}

	isPrefix = len(node.children) > 0
	return lastMatch, lastMatchLen, isPrefix
}
