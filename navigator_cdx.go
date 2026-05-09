package foxy

import (
	"io"

	"github.com/mkfoss/foxy/pkg/core"
)

// CdxNavigator navigates a DBF file in index order using a CDX index.
// Unlike DefaulNavigator which navigates records in physical order,
// CdxNavigator uses a CDX index tag to navigate records in sorted order.
type CdxNavigator struct {
	position     int
	offset       int
	count        int
	size         int
	rdr          io.ReadSeeker
	readfunc     ReadFunc
	dbf          *Dbf          // Reference to parent Dbf for CDX access
	tagName      string        // Name of the CDX tag to use
	iterator     *core.CdxIterator
	curRecNo     int32         // Current record number from iterator
	isDescending bool          // True if this is a descending index
}

// NewCdxNavigator creates a new CDX navigator for the specified tag name.
func NewCdxNavigator(tagName string) *CdxNavigator {
	return &CdxNavigator{
		tagName: tagName,
	}
}


// Initialize sets up the navigator with DBF file information
func (nav *CdxNavigator) Initialize(readseeker io.ReadSeeker, offset, size, count int, rf ReadFunc) error {
	nav.rdr = readseeker
	nav.offset = offset
	nav.size = size
	nav.count = count
	nav.readfunc = rf
	nav.position = 0
	nav.curRecNo = 0

	return nil
}

// SetDbf sets the parent Dbf reference (called by Dbf.SetNavigator)
func (nav *CdxNavigator) SetDbf(dbf *Dbf) {
	nav.dbf = dbf
}

// ensureIterator ensures the CDX file is loaded and iterator is initialized
func (nav *CdxNavigator) ensureIterator() error {
	if nav.iterator != nil {
		return nil
	}

	// Ensure CDX is loaded
	if err := nav.dbf.ensureCdxLoaded(); err != nil {
		return NewNavigationError().SetContext("cdx navigator: ensure cdx loaded").SetWrapped(err)
	}

	// Find the specified tag
	tag := nav.dbf.cdx.FindTag(nav.tagName)
	if tag == nil {
		return NewNavigationError().SetContext("cdx navigator: tag not found: " + nav.tagName)
	}

	// Create iterator
	nav.iterator = tag.NewIterator(nav.dbf.cdx)
	nav.isDescending = tag.Descending

	return nil
}

// Finalize cleans up the navigator
func (nav *CdxNavigator) Finalize() error {
	nav.rdr = nil
	nav.readfunc = nil
	nav.dbf = nil
	nav.iterator = nil
	nav.offset = 0
	nav.size = 0
	nav.count = 0
	nav.position = 0
	nav.curRecNo = 0

	return nil
}

// First positions at the first record in index order
func (nav *CdxNavigator) First() error {
	if err := nav.ensureIterator(); err != nil {
		return err
	}

	// For descending indexes, "first" means the largest value (Last in B-tree)
	// For ascending indexes, "first" means the smallest value (First in B-tree)
	var recNo int32
	var ok bool

	if nav.isDescending {
		recNo, ok = nav.iterator.Last()
	} else {
		recNo, ok = nav.iterator.First()
	}

	if !ok {
		return NewNavigationEofError().SetContext("cdx navigator first")
	}

	nav.curRecNo = recNo
	nav.position = 1
	return nav.readRecord(recNo)
}

// Last positions at the last record in index order
func (nav *CdxNavigator) Last() error {
	if err := nav.ensureIterator(); err != nil {
		return err
	}

	// For descending indexes, "last" means the smallest value (First in B-tree)
	// For ascending indexes, "last" means the largest value (Last in B-tree)
	var recNo int32
	var ok bool

	if nav.isDescending {
		recNo, ok = nav.iterator.First()
	} else {
		recNo, ok = nav.iterator.Last()
	}

	if !ok {
		return NewNavigationEofError().SetContext("cdx navigator last")
	}

	nav.curRecNo = recNo
	nav.position = nav.count
	return nav.readRecord(recNo)
}

// Next moves to the next record in index order
func (nav *CdxNavigator) Next() error {
	if nav.position >= nav.count {
		return NewNavigationEofError().SetContext("cdx navigator next")
	}

	if err := nav.ensureIterator(); err != nil {
		return err
	}

	// For descending indexes, "next" means move toward smaller values (Previous in B-tree)
	// For ascending indexes, "next" means move toward larger values (Next in B-tree)
	var recNo int32
	var ok bool

	if nav.isDescending {
		recNo, ok = nav.iterator.Previous()
	} else {
		recNo, ok = nav.iterator.Next()
	}

	if !ok {
		return NewNavigationEofError().SetContext("cdx navigator next")
	}

	nav.curRecNo = recNo
	nav.position++
	return nav.readRecord(recNo)
}

// Previous moves to the previous record in index order
func (nav *CdxNavigator) Previous() error {
	if nav.position <= 1 {
		return NewNavigationBofError().SetContext("cdx navigator previous")
	}

	if err := nav.ensureIterator(); err != nil {
		return err
	}

	// For descending indexes, "previous" means move toward larger values (Next in B-tree)
	// For ascending indexes, "previous" means move toward smaller values (Previous in B-tree)
	var recNo int32
	var ok bool

	if nav.isDescending {
		recNo, ok = nav.iterator.Next()
	} else {
		recNo, ok = nav.iterator.Previous()
	}

	if !ok {
		return NewNavigationBofError().SetContext("cdx navigator previous")
	}

	nav.curRecNo = recNo
	nav.position--
	return nav.readRecord(recNo)
}

// Position returns the current 1-based position in index order
func (nav *CdxNavigator) Position() int {
	return nav.position
}

// Goto moves to a specific 1-based position in index order
// Note: For large datasets, this is expensive as it requires iteration from First()
func (nav *CdxNavigator) Goto(position int) error {
	if position < 1 || position > nav.count {
		return NewNavigationError().SetContext("cdx navigator goto: invalid position")
	}

	if err := nav.ensureIterator(); err != nil {
		return err
	}

	// Start from the beginning and iterate to the desired position
	err := nav.First()
	if err != nil {
		return err
	}

	// Iterate to the target position
	for i := 1; i < position; i++ {
		err = nav.Next()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsFirst reports whether the current position is the first record
func (nav *CdxNavigator) IsFirst() bool {
	return nav.position == 1
}

// IsLast reports whether the current position is the last record
func (nav *CdxNavigator) IsLast() bool {
	return nav.position == nav.count
}

// readRecord reads the record with the given record number
func (nav *CdxNavigator) readRecord(recNo int32) error {
	// Calculate the byte offset in the DBF file
	// recNo is 1-based, so subtract 1
	byteOffset := int64(nav.offset + nav.size*int(recNo-1))

	// Seek to the record position
	_, err := nav.rdr.Seek(byteOffset, io.SeekStart)
	if err != nil {
		return NewNavigationError().SetContext("cdx navigator: seek failed").SetWrapped(err)
	}

	return nav.readfunc()
}
