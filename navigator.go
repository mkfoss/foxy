package foxy

import "io"

// ReadFunc is a function passed to Navigator.Initialize from Dbf that reads data into the current record buffer, Navigator
// should call the func as needed during the navigation.
type ReadFunc = func() error

type Navigator interface {
	// First - Navigate to the first record and read (implementor is responsible for calling ReadFunc)
	First() error
	// Next -  Navigate to the next record and read (implementor is responsible for calling ReadFunc)
	Next() error
	// Previous - Navigate to the previous record and read (implementor is responsible for calling ReadFunc)
	Previous() error
	// Last - Navigate to the last record and read (implementor is responsible for calling ReadFunc)
	Last() error
	// Goto - Navigate to an arbitrary record and read (implementor is responsible for calling ReadFunc)
	Goto(position int) error
	// Position - current record index
	Position() int
	// Initialize - initialize the navigator with the neccessary info
	Initialize(readseeker io.ReadSeeker, offset, size, count int, rf ReadFunc) error
	// Finalize - clean up any resources on closing
	Finalize() error
	// IsFirst - report whether the current position is the first record
	IsFirst() bool
	// IsLast - report whether the current position is the last record
	IsLast() bool
}
