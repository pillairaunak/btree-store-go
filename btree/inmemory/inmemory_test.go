// btree/inmemory/inmemory_test.go
package inmemory

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pillairaunak/btree-store-go/btree"
)

func TestInMemoryBTree_Lookup(t *testing.T) {
	tree := NewInMemoryBTree()

	t.Run("LookupNonExistentKey", func(t *testing.T) {
		_, found := tree.Lookup(42)
		if found {
			t.Errorf("Expected key 42 to not be found, but it was")
		}
	})

	t.Run("InsertAndLookup", func(t *testing.T) {
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
	})

	t.Run("UpdateAndLookup", func(t *testing.T) {
		if err := tree.Insert(42, 200); err != nil {
			t.Fatalf("Insert failed: %v", err)

		}
		value, found := tree.Lookup(42)
		if !found {
			t.Errorf("Expected key 42 to be found, but it wasn't")
		}
		if value != 200 {
			t.Errorf("Expected updated value 200 for key 42, got %d", value)
		}
	})
}

func TestInMemoryBTree_Insert(t *testing.T) {
	tree := NewInMemoryBTree()

	t.Run("InsertNewKey", func(t *testing.T) {
		err := tree.Insert(1, 100)
		if err != nil {
			t.Errorf("Expected Insert to succeed, but got error: %v", err)
		}
	})

	t.Run("UpdateExistingKey", func(t *testing.T) {
		err := tree.Insert(1, 200)
		if err != nil {
			t.Errorf("Expected Insert to succeed, but got error: %v", err)
		}

		// Verify the update worked
		value, _ := tree.Lookup(1)
		if value != 200 {
			t.Errorf("Expected value 200 after update, got %d", value)
		}
	})
}
func TestInMemoryBTree_Scan(t *testing.T) {
	tree := NewInMemoryBTree()

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

	// Helper function to convert channel results to a slice
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

	// --- Test Cases ---
	t.Run("ScanEntireRange", func(t *testing.T) {
		resultsChan, err := tree.Scan(0, 100)
		results, err := collectResults(resultsChan, err)

		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}
	})

	t.Run("ScanPartialRange", func(t *testing.T) {
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

	t.Run("ScanEmptyRange", func(t *testing.T) {
		resultsChan, err := tree.Scan(60, 70)
		results, err := collectResults(resultsChan, err)

		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty range, got %d", len(results))
		}
	})

	t.Run("MinKeyGreaterThanMaxKey", func(t *testing.T) {
		resultsChan, err := tree.Scan(70, 60)
		results, err := collectResults(resultsChan, err)
		if err != nil {
			t.Errorf("Scan returned unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results when minKey > maxKey, got %d", len(results))
		}
	})

	t.Run("ScanWithBoundaries", func(t *testing.T) {
		resultsChan, err := tree.Scan(10, 50) // Exact boundaries
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
