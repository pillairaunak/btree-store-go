// buffermanager/buffermanager.go
package buffermanager

import (
	"errors"
	"fmt"
	"github.com/pillairaunak/btree-store-go/btree" // Import the btree interface
	"github.com/pillairaunak/btree-store-go/btree/inmemory"
)

// Common errors that might occur during buffer manager operations
var (
	ErrBTreeNotFound = errors.New("btree not found")
	ErrPageNotFound  = errors.New("page not found")
	ErrBufferFull    = errors.New("buffer is full")
)

// PageID uniquely identifies a page within a BTree
type PageID uint64

// BufferManager defines the interface for managing the buffer pool and BTrees.
type BufferManager interface {
	// CreateBTree creates a new empty BTree and returns its identifier.
	CreateBTree() (string, error)

	// OpenBTree opens an existing BTree by its identifier.
	// Returns a BTree interface representing the opened B-Tree.
	OpenBTree(btreeID string) (btree.BTree, error)

	// DeleteBTree permanently removes a BTree.
	DeleteBTree(btreeID string) error

	// CloseBTree closes an open BTree.
	CloseBTree(btreeID string) error

	// PinPage loads a page into the buffer pool and pins it, preventing eviction.
	// Returns the page data and its position in the buffer.
	PinPage(btreeID string, pageID PageID) ([]byte, int, error)

	// UnpinPage marks a page as unpinned, making it eligible for eviction.
	// The dirty parameter indicates whether the page was modified.
	UnpinPage(bufferPos int, dirty bool) error

	// AllocatePage creates a new page for a BTree and returns its PageID.
	AllocatePage(btreeID string) (PageID, error)

	// FreePage marks a page as free for future allocation.
	FreePage(btreeID string, pageID PageID) error
}

// Option represents a configuration option for the buffer manager.
type Option func(*bufferManagerConfig)

// WithDirectory specifies the directory where BTree files are stored.
func WithDirectory(dir string) Option {
	return func(config *bufferManagerConfig) {
		config.directory = dir
	}
}

// WithBufferSize specifies the maximum number of pages to keep in memory.
func WithBufferSize(pages int) Option {
	return func(config *bufferManagerConfig) {
		config.bufferSize = pages
	}
}

// bufferManagerConfig holds the internal configuration for the buffer manager.
type bufferManagerConfig struct {
	directory  string
	bufferSize int
}

// mockBufferManager implements the BufferManager interface for testing.
type mockBufferManager struct {
	btrees      map[string]btree.BTree // Map BTreeID to BTree interface
	pages       map[string]map[PageID][]byte
	buffer      map[int]bufferEntry
	nextBTreeID int
	nextPageID  map[string]PageID
	config      bufferManagerConfig
}

// bufferEntry represents a page in the buffer pool.
type bufferEntry struct {
	btreeID string
	pageID  PageID
	data    []byte
	pinned  bool
	dirty   bool
}

// NewMockBufferManager creates a new mock buffer manager with optional parameters.
func NewMockBufferManager(options ...Option) *mockBufferManager {
	config := bufferManagerConfig{
		directory:  ".", // Default directory
		bufferSize: 10,  // Default buffer size
	}

	for _, option := range options {
		option(&config)
	}

	return &mockBufferManager{
		btrees:      make(map[string]btree.BTree),
		pages:       make(map[string]map[PageID][]byte),
		buffer:      make(map[int]bufferEntry),
		nextBTreeID: 1,
		nextPageID:  make(map[string]PageID),
		config:      config,
	}
}

// CreateBTree creates a new empty BTree and returns its identifier.
func (m *mockBufferManager) CreateBTree() (string, error) {
	btreeID := fmt.Sprintf("btree_%d", m.nextBTreeID)
	m.nextBTreeID++
	//For Mock Implementation, We are creating a new in memory Btree and saving.
	m.btrees[btreeID] = inmemory.NewInMemoryBTree()
	m.pages[btreeID] = make(map[PageID][]byte)
	m.nextPageID[btreeID] = 1
	return btreeID, nil
}

// OpenBTree opens an existing BTree by its identifier.
func (m *mockBufferManager) OpenBTree(btreeID string) (btree.BTree, error) {
	b, exists := m.btrees[btreeID]
	if !exists {
		return nil, ErrBTreeNotFound
	}
	return b, nil // Return the BTree interface
}

// DeleteBTree permanently removes a BTree.
func (m *mockBufferManager) DeleteBTree(btreeID string) error {
	if _, exists := m.btrees[btreeID]; !exists {
		return ErrBTreeNotFound
	}
	delete(m.btrees, btreeID)
	delete(m.pages, btreeID)
	delete(m.nextPageID, btreeID)

	// Clean up any buffer entries associated with this B-Tree
	for pos, entry := range m.buffer {
		if entry.btreeID == btreeID {
			delete(m.buffer, pos)
		}
	}
	return nil
}

// CloseBTree closes an open BTree.
func (m *mockBufferManager) CloseBTree(btreeID string) error {
	if _, exists := m.btrees[btreeID]; !exists {
		return ErrBTreeNotFound
	}

	// Check for pinned pages.  In a real implementation, we would likely
	// want to either return an error or force-flush pinned pages.
	for _, entry := range m.buffer {
		if entry.btreeID == btreeID && entry.pinned {
			return fmt.Errorf("cannot close BTree %s: pages still pinned", btreeID)
		}
	}

	//Flush all the dirty pages before closing
	for pos, entry := range m.buffer {
		if entry.btreeID == btreeID {
			if entry.dirty {
				m.pages[entry.btreeID][entry.pageID] = entry.data
			}
			delete(m.buffer, pos)
		}
	}
	return nil
}

// PinPage loads a page into the buffer pool and pins it.
func (m *mockBufferManager) PinPage(btreeID string, pageID PageID) ([]byte, int, error) {
	if _, exists := m.btrees[btreeID]; !exists {
		return nil, 0, ErrBTreeNotFound
	}

	pageData, exists := m.pages[btreeID][pageID]
	if !exists {
		return nil, 0, ErrPageNotFound
	}

	// Find a free buffer position (simplified - no eviction yet)
	bufferPos := -1
	for i := 0; i < m.config.bufferSize; i++ {
		if _, exists := m.buffer[i]; !exists {
			bufferPos = i
			break
		}
	}
	if bufferPos == -1 {
		return nil, 0, ErrBufferFull // Simplified - no eviction
	}

	m.buffer[bufferPos] = bufferEntry{
		btreeID: btreeID,
		pageID:  pageID,
		data:    pageData,
		pinned:  true,
		dirty:   false,
	}

	return pageData, bufferPos, nil
}

// UnpinPage marks a page as unpinned.
func (m *mockBufferManager) UnpinPage(bufferPos int, dirty bool) error {
	entry, exists := m.buffer[bufferPos]
	if !exists {
		return ErrPageNotFound
	}

	if !entry.pinned {
		return fmt.Errorf("page at buffer position %d is not pinned", bufferPos)
	}

	entry.pinned = false
	if dirty {
		entry.dirty = true

		m.pages[entry.btreeID][entry.pageID] = entry.data
	}
	m.buffer[bufferPos] = entry

	return nil
}

// AllocatePage creates a new page for a BTree.
func (m *mockBufferManager) AllocatePage(btreeID string) (PageID, error) {
	if _, exists := m.btrees[btreeID]; !exists {
		return 0, ErrBTreeNotFound
	}

	pageID := m.nextPageID[btreeID]
	m.nextPageID[btreeID]++
	m.pages[btreeID][pageID] = make([]byte, 4096) // 4KB page size
	return pageID, nil
}

// FreePage marks a page as free.
func (m *mockBufferManager) FreePage(btreeID string, pageID PageID) error {
	if _, exists := m.btrees[btreeID]; !exists {
		return ErrBTreeNotFound
	}

	if _, exists := m.pages[btreeID][pageID]; !exists {
		return ErrPageNotFound
	}

	// Remove from buffer if present (important for consistency)
	for pos, entry := range m.buffer {
		if entry.btreeID == btreeID && entry.pageID == pageID {
			delete(m.buffer, pos)
			break
		}
	}

	delete(m.pages[btreeID], pageID)
	return nil
}
