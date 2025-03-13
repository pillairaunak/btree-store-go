// btree/btree_test.go
package btree_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pillairaunak/btree-store-go/btree"
	"github.com/pillairaunak/btree-store-go/btree/inmemory" // Import the inmemory implementation
)

/*func TestBTreeInterface(t *testing.T) {
	// Use the InMemoryBTree as a concrete implementation for testing the interface.
	tree := inmemory.NewInMemoryBTree()

	t.Run("Lookup", func(t *testing.T) {
		testBTreeLookup(t, tree)
	})

	t.Run("Insert", func(t *testing.T) {
		testBTreeInsert(t, tree)
	})

	t.Run("Scan", func(t *testing.T) {
		testBTreeScan(t, tree)
	})
}*/

func TestBTreeInterface(t *testing.T) {
	t.Run("Lookup", func(t *testing.T) {
		// Use a fresh tree for each test
		tree := inmemory.NewInMemoryBTree()
		testBTreeLookup(t, tree)
	})

	t.Run("Insert", func(t *testing.T) {
		// Use a fresh tree for each test
		tree := inmemory.NewInMemoryBTree()
		testBTreeInsert(t, tree)
	})

	t.Run("Scan", func(t *testing.T) {
		// Use a fresh tree for each test
		tree := inmemory.NewInMemoryBTree()
		testBTreeScan(t, tree)
	})
}

func testBTreeLookup(t *testing.T, tree btree.BTree) {
	// Test case 1: Looking up a non-existent key
	_, found := tree.Lookup(42)
	if found {
		t.Errorf("Expected key 42 to not be found, but it was")
	}

	// Test case 2: Insert and lookup
	if err := tree.Insert(42, 100); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	value, found := tree.Lookup(42)
	if !found {
		t.Errorf("Expected key 42 to be found, but it wasn't")
	}
	if value != 100 {
		t.Errorf("Expected value 100 for key 42, got %d", value)
	}

	// Test case 3: Update and lookup
	if err := tree.Insert(42, 200); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	value, found = tree.Lookup(42)
	if !found {
		t.Errorf("Expected key 42 to be found, but it wasn't")
	}
	if value != 200 {
		t.Errorf("Expected updated value 200 for key 42, got %d", value)
	}
}

func testBTreeInsert(t *testing.T, tree btree.BTree) {
	// Test case 1: Insert new key
	err := tree.Insert(1, 100)
	if err != nil {
		t.Errorf("Expected Insert to succeed, but got error: %v", err)
	}

	// Test case 2: Update existing key
	err = tree.Insert(1, 200)
	if err != nil {
		t.Errorf("Expected Insert to succeed, but got error: %v", err)
	}

	// Verify the update worked
	value, _ := tree.Lookup(1)
	if value != 200 {
		t.Errorf("Expected value 200 after update, got %d", value)
	}
}

func testBTreeScan(t *testing.T, tree btree.BTree) {
	// Insert test data
	testData := map[uint64]uint64{
		10: 100,
		20: 200,
		30: 300,
		40: 400,
		50: 500,
	}

	for k, v := range testData {
		if err := tree.Insert(k, v); err != nil {
			t.Fatalf("Insert error %v", err)
		}
	}

	// Helper function to convert channel results to a slice for easier comparison
	collectResults := func(results <-chan btree.KeyValuePair, err error) ([]btree.KeyValuePair, error) {
		if err != nil {
			return nil, err
		}
		var collected []btree.KeyValuePair
		for r := range results {
			collected = append(collected, r)
		}
		return collected, nil
	}

	// Test case 1: Scan entire range
	t.Run("Scan entire range", func(t *testing.T) {
		resultsChan, err := tree.Scan(0, 100)
		results, err := collectResults(resultsChan, err)

		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
	})
	// Test case 2: Scan partial range
	t.Run("Scan partial range", func(t *testing.T) {
		resultsChan, err := tree.Scan(20, 40)
		results, err := collectResults(resultsChan, err)

		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results for range [20, 40], got %d", len(results))
		}
		// Ensure results are sorted
		if !sort.SliceIsSorted(results, func(i, j int) bool {
			return results[i].Key < results[j].Key
		}) {
			t.Error("Scan results are not sorted by key")
		}
	})

	// Test case 3: Scan empty range
	t.Run("Scan empty range", func(t *testing.T) {
		resultsChan, err := tree.Scan(60, 70)
		results, err := collectResults(resultsChan, err)

		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty range, got %d", len(results))
		}
	})

	// Test case 4: minKey > maxKey
	t.Run("minKey > maxKey", func(t *testing.T) {
		resultsChan, err := tree.Scan(70, 60)
		results, err := collectResults(resultsChan, err)
		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results when minKey > maxKey, got %d", len(results))
		}
	})

	// Test case 5: Scan with boundaries
	t.Run("Scan with boundaries", func(t *testing.T) {
		resultsChan, err := tree.Scan(10, 50) // Exact boundaries of the data
		results, err := collectResults(resultsChan, err)
		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results for range [10, 50], got %d", len(results))
		}

		expected := []btree.KeyValuePair{
			{Key: 10, Value: 100},
			{Key: 20, Value: 200},
			{Key: 30, Value: 300},
			{Key: 40, Value: 400},
			{Key: 50, Value: 500},
		}

		if !reflect.DeepEqual(results, expected) {
			t.Errorf("Scan results mismatch.\nExpected: %v\nGot: %v", expected, results)
		}
	})
}
