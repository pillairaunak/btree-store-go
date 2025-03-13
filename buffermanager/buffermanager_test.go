// buffermanager/buffermanager_test.go
package buffermanager

import (
	"testing"
)

func TestBufferManager_CreateAndDeleteBTree(t *testing.T) {
	bm := NewMockBufferManager()

	t.Run("CreateBTree", func(t *testing.T) {
		btreeID, err := bm.CreateBTree()
		if err != nil {
			t.Fatalf("CreateBTree failed: %v", err)
		}
		if btreeID == "" {
			t.Fatal("CreateBTree returned empty BTreeID")
		}

		_, err = bm.OpenBTree(btreeID)
		if err != nil {
			t.Fatalf("BTree with id %s not properly created: %v", btreeID, err)
		}
	})

	t.Run("DeleteBTree", func(t *testing.T) {
		btreeID, _ := bm.CreateBTree()
		err := bm.DeleteBTree(btreeID)
		if err != nil {
			t.Fatalf("DeleteBTree failed: %v", err)
		}

		err = bm.DeleteBTree(btreeID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}

		_, err = bm.OpenBTree(btreeID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})

	t.Run("DeleteNonExistentBTree", func(t *testing.T) {
		err := bm.DeleteBTree("nonexistent")
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})
}

func TestBufferManager_OpenBTree(t *testing.T) {
	bm := NewMockBufferManager()

	t.Run("OpenNonExistentBTree", func(t *testing.T) {
		_, err := bm.OpenBTree("nonexistent")
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})

	t.Run("OpenExistingBTree", func(t *testing.T) {
		btreeID, _ := bm.CreateBTree()
		bTree, err := bm.OpenBTree(btreeID)
		if err != nil {
			t.Fatalf("OpenBTree failed: %v", err)
		}
		if bTree == nil {
			t.Fatal("OpenBTree returned nil BTree")
		}
	})
}

func TestBufferManager_CloseBTree(t *testing.T) {
	bm := NewMockBufferManager()

	t.Run("Close a Btree", func(t *testing.T) {
		btreeID, err := bm.CreateBTree()
		if err != nil {
			t.Fatalf("CreateBtree failed: %v", err)
		}

		err = bm.CloseBTree(btreeID)
		if err != nil {
			t.Fatalf("CloseBtree failed: %v", err)
		}
	})
	t.Run("Closing non existing Btree", func(t *testing.T) {
		btreeID := "btee_1000"

		err := bm.CloseBTree(btreeID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})

	t.Run("Close BTree with pinned pages", func(t *testing.T) {
		bm := NewMockBufferManager()
		btreeID, _ := bm.CreateBTree()
		pageID, _ := bm.AllocatePage(btreeID)

		_, _, _ = bm.PinPage(btreeID, pageID)

		err := bm.CloseBTree(btreeID)
		if err == nil {
			t.Fatal("Expected an error when closing BTree with pinned pages")
		}
	})

}

func TestBufferManager_PageOperations(t *testing.T) {
	bm := NewMockBufferManager()
	btreeID, _ := bm.CreateBTree()

	t.Run("AllocatePage", func(t *testing.T) {
		pageID, err := bm.AllocatePage(btreeID)
		if err != nil {
			t.Fatalf("AllocatePage failed: %v", err)
		}
		if pageID == 0 {
			t.Fatal("AllocatePage returned invalid PageID")
		}
	})

	t.Run("PinPage", func(t *testing.T) {
		pageID, _ := bm.AllocatePage(btreeID)
		data, bufferPos, err := bm.PinPage(btreeID, pageID)
		if err != nil {
			t.Fatalf("PinPage failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("PinPage returned empty data")
		}
		if bufferPos < 0 {
			t.Error("PinPage returned invalid buffer position")
		}
	})

	t.Run("UnpinPage", func(t *testing.T) {
		pageID, _ := bm.AllocatePage(btreeID)
		_, bufferPos, _ := bm.PinPage(btreeID, pageID)
		err := bm.UnpinPage(bufferPos, true)
		if err != nil {
			t.Fatalf("UnpinPage failed: %v", err)
		}
	})
	t.Run("Unpin Not pinned", func(t *testing.T) {
		bufferPos := 1000
		err := bm.UnpinPage(bufferPos, true)
		if err == nil {
			t.Fatalf("UnpinPage on unpinned buffer position didn't fail")
		}
	})

	t.Run("PinNonExistentPage", func(t *testing.T) {
		_, _, err := bm.PinPage(btreeID, 999)
		if err != ErrPageNotFound {
			t.Fatalf("Expected ErrPageNotFound, got: %v", err)
		}
	})

	t.Run("AllocateAndFreePage", func(t *testing.T) {
		pageID, _ := bm.AllocatePage(btreeID)
		err := bm.FreePage(btreeID, pageID)
		if err != nil {
			t.Fatalf("FreePage failed: %v", err)
		}

		_, _, err = bm.PinPage(btreeID, pageID)
		if err != ErrPageNotFound {
			t.Fatalf("Expected ErrPageNotFound after FreePage, got: %v", err)
		}
	})
	t.Run("Free a Page", func(t *testing.T) {

		pageID, err := bm.AllocatePage(btreeID)
		if err != nil {
			t.Fatalf("AllocatePage failed: %v", err)
		}
		if pageID == 0 {
			t.Fatal("AllocatePage returned invalid PageID")
		}

		err = bm.FreePage(btreeID, pageID)
		if err != nil {
			t.Fatalf("FreePage failed: %v", err)
		}
	})
	t.Run("Free a non allocated page", func(t *testing.T) {
		pageID := PageID(1000)

		err := bm.FreePage(btreeID, pageID)
		if err != ErrPageNotFound {
			t.Fatalf("Expected ErrPageNotFound, got : %v", err)
		}
	})
	t.Run("PinPage on a deleted BTree", func(t *testing.T) {
		btreeID, _ := bm.CreateBTree()
		pageID, _ := bm.AllocatePage(btreeID)
		bm.DeleteBTree(btreeID)

		_, _, err := bm.PinPage(btreeID, pageID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})
	t.Run("AllocatePage on a deleted BTree", func(t *testing.T) {
		btreeID, _ := bm.CreateBTree()
		bm.DeleteBTree(btreeID)

		_, err := bm.AllocatePage(btreeID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})
	t.Run("FreePage on a deleted BTree", func(t *testing.T) {
		btreeID, _ := bm.CreateBTree()
		pageID, _ := bm.AllocatePage(btreeID)
		bm.DeleteBTree(btreeID)

		err := bm.FreePage(btreeID, pageID)
		if err != ErrBTreeNotFound {
			t.Fatalf("Expected ErrBTreeNotFound, got: %v", err)
		}
	})
}
