// btree/btree.go
package btree

// BTree defines the interface for a B-Tree data structure
// that stores uint64 keys and values.
type BTree interface {
	// Lookup finds the value associated with the given key.
	// Returns the value and true if found, or 0 and false if not found.
	Lookup(key uint64) (value uint64, found bool)

	// Insert adds or updates a key-value pair in the tree.
	// Returns an error if the operation fails, nil otherwise.
	Insert(key uint64, value uint64) error

	// Scan retrieves all key-value pairs where the key is between minKey and maxKey (inclusive).
	// Results are streamed via a channel in ascending key order.
	// The channel is closed after the last result or if an error occurs.
	Scan(minKey uint64, maxKey uint64) (<-chan KeyValuePair, error)
}

// KeyValuePair represents a key-value pair in the B+Tree
type KeyValuePair struct {
	Key   uint64
	Value uint64
}
