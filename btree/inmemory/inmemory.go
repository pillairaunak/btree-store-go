// btree/inmemory/inmemory.go
package inmemory

import (
	"github.com/pillairaunak/btree-store-go/btree" // Import the btree interface
	"sort"
)

// InMemoryBTree implements the BTree interface with an in-memory map.
type InMemoryBTree struct {
	Data map[uint64]uint64
}

// NewInMemoryBTree creates a new instance of the in-memory BTree.
func NewInMemoryBTree() *InMemoryBTree {
	return &InMemoryBTree{
		Data: make(map[uint64]uint64),
	}
}

// Lookup finds the value associated with the given key.
func (m *InMemoryBTree) Lookup(key uint64) (uint64, bool) {
	value, found := m.Data[key]
	return value, found
}

// Insert adds or updates a key-value pair in the tree.
func (m *InMemoryBTree) Insert(key uint64, value uint64) error {
	m.Data[key] = value
	return nil
}

// Scan retrieves all key-value pairs within the given range.
func (m *InMemoryBTree) Scan(minKey uint64, maxKey uint64) (<-chan btree.KeyValuePair, error) {
	results := make(chan btree.KeyValuePair)

	go func() {
		defer close(results)

		// Create a slice of keys for sorting
		var keys []uint64
		for k := range m.Data {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		for _, k := range keys {
			if k >= minKey && k <= maxKey {
				results <- btree.KeyValuePair{Key: k, Value: m.Data[k]}
			}
		}
	}()

	return results, nil
}
