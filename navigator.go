package foxy

import "io"

// ReadFunc is a function passed to Navigator.Initialize from Dbf that reads data into the current record buffer.
// Navigator implementations should call this function whenever they position the file pointer at a new record.
type ReadFunc = func() error

// Navigator defines an interface for traversing records in a DBF file.
// Different implementations can provide different traversal orders (e.g., physical, indexed).
type Navigator interface {
	// First positions the file at the first record and reads it.
	First() error
	// Next positions the file at the next record and reads it.
	Next() error
	// Previous positions the file at the previous record and reads it.
	Previous() error
	// Last positions the file at the last record and reads it.
	Last() error
	// Goto positions the file at an arbitrary record (1-based) and reads it.
	Goto(position int) error
	// Position returns the current 1-based record index.
	Position() int
	// Initialize sets up the navigator with the necessary file and table information.
	Initialize(readseeker io.ReadSeeker, offset, size, count int, rf ReadFunc) error
	// Finalize cleans up any resources used by the navigator.
	Finalize() error
	// IsFirst reports whether the current position is the first record.
	IsFirst() bool
	// IsLast reports whether the current position is the last record.
	IsLast() bool
}
