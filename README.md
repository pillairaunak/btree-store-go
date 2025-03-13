# B-Tree Store

A modular B-Tree based key-value store implemented in Go, designed for flexibility and experimentation with different B-Tree variants.

## Overview

This project implements a key-value store using B-Tree data structures in Go. The system emphasizes modularity, testability, and adherence to Go best practices, allowing for easy extension with different B-Tree variants.

## Architecture

The system consists of two primary components:

1. **B-Tree Interface and Implementations**: Defines an abstract interface for all B-Tree variants and provides concrete implementations.
2. **Buffer Manager**: Manages the interaction between in-memory B-Tree structures and persistent storage.

The core design principle is the use of Go's interfaces to decouple components, allowing:
- Easy addition of new B-Tree variants without modifying the Buffer Manager
- Experimentation with different B-Tree algorithms
- Performance comparison between implementations
- Thorough testing of each component in isolation

## Key Components

### B-Tree Interface

The `BTree` interface defines the contract for all B-Tree implementations:

```go
type BTree interface {
    Lookup(key uint64) (value uint64, found bool)
    Insert(key uint64, value uint64) error
    Scan(minKey uint64, maxKey uint64) (<-chan KeyValuePair, error)
}
```

### B-Tree Implementations

Each B-Tree variant is implemented in its own package:

- `btree/inmemory`: An in-memory B-Tree implementation for testing and scenarios where persistence isn't required
- `btree/bplustree` (Future): A B+Tree implementation
- `btree/b*tree` (Future): A B*Tree implementation

### Buffer Manager

The `BufferManager` interface handles interaction with persistent storage:

```go
type BufferManager interface {
    CreateBTree() (string, error)
    OpenBTree(btreeID string) (btree.BTree, error)
    DeleteBTree(btreeID string) error
    CloseBTree(btreeID string) error
    PinPage(btreeID string, pageID PageID) ([]byte, int, error)
    UnpinPage(bufferPos int, dirty bool) error
    AllocatePage(btreeID string) (PageID, error)
    FreePage(btreeID string, pageID PageID) error
}
```

## Project Structure

```
btree-store-go/
├── btree/
│   ├── btree.go           // BTree interface
│   ├── btree_test.go      // BTree interface tests
│   ├── inmemory/          // In-memory implementation
│   │   ├── inmemory.go
│   │   └── inmemory_test.go
│   ├── bplustree/         // Future B+Tree package
│   │   ├── bplustree.go
│   │   └── bplustree_test.go
│   └── ...                // Other B-Tree variants
└── buffermanager/
    ├── buffermanager.go   // BufferManager interface and mock
    └── buffermanager_test.go // BufferManager tests
```

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Installation

```bash
git clone https://github.com/yourusername/btree-store-go.git
cd btree-store-go
go test ./...
```

## Testing

The project employs a comprehensive testing strategy:

1. **Interface Tests**: Verify that implementations satisfy the `btree.BTree` interface
2. **Unit Tests**: Focus on specific implementations of each B-Tree variant
3. **Integration Tests**: Verify interactions between the Buffer Manager and B-Tree implementations

Run all tests with:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Extensibility

Adding a new B-Tree variant involves:

1. Creating a new package (e.g., `btree/bplustree`)
2. Defining a struct to hold the B-Tree's data
3. Implementing the `btree.BTree` interface methods
4. Writing unit tests for the new implementation

## License

[MIT License](LICENSE)
