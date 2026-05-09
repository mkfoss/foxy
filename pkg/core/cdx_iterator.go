package core

import (
	"encoding/binary"
	"io"
)

// CdxIterator provides ordered iteration through index keys
type CdxIterator struct {
	tag            *CdxTag
	cdxFile        *CdxFile
	curRecNo       int32
	curBlockOffset int64   // Current leaf block
	curKeyIndex    int     // Index within current leaf block
	leafRecords    []int32 // Cached records from current leaf
	eof            bool
	bof            bool
}

// NewIterator creates a new iterator for the given tag
func (tag *CdxTag) NewIterator(cdxFile *CdxFile) *CdxIterator {
	return &CdxIterator{
		tag:     tag,
		cdxFile: cdxFile,
		eof:     false,
		bof:     true,
	}
}

// First positions at the first key in index order
// Returns the record number and true if successful, 0 and false if empty
func (it *CdxIterator) First() (int32, bool) {
	if it.tag == nil || it.cdxFile == nil {
		it.eof = true
		it.bof = true
		return 0, false
	}

	// Seek to find the leftmost leaf block
	recNo, result := it.seekToFirst()
	if result == SeekError || result == SeekEOF {
		it.eof = true
		it.bof = true
		it.curRecNo = 0
		return 0, false
	}

	it.eof = false
	it.bof = false
	it.curRecNo = recNo
	return recNo, true
}

// seekToFirst finds the first (leftmost) record in the index
func (it *CdxIterator) seekToFirst() (int32, SeekResult) {
	// Start from root and always go left/first
	return it.seekToExtreme(int64(it.tag.RootBlock), true)
}

// seekToExtreme recursively navigates to the leftmost (first=true) or rightmost (first=false) record
func (it *CdxIterator) seekToExtreme(blockOffset int64, first bool) (int32, SeekResult) {
	// Read the block
	_, err := it.cdxFile.File.Seek(blockOffset, io.SeekStart)
	if err != nil {
		return 0, SeekError
	}

	blockBuf := make([]byte, CDXBlockSize)
	n, err := it.cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		return 0, SeekError
	}

	// Parse block header
	blockType := binary.LittleEndian.Uint16(blockBuf[0:2])
	numKeys := binary.LittleEndian.Uint16(blockBuf[2:4])

	if numKeys == 0 {
		return 0, SeekEOF
	}

	isLeaf := (blockType & 0x03) > 1

	if isLeaf {
		// Leaf block - extract all records and populate iterator state
		records := extractCompactLeafRecords(blockBuf, int(numKeys))
		if len(records) == 0 {
			return 0, SeekError
		}

		// Populate iterator state for Next/Previous navigation
		it.curBlockOffset = blockOffset
		it.leafRecords = records

		if first {
			it.curKeyIndex = 0
			return records[0], SeekSuccess
		} else {
			it.curKeyIndex = len(records) - 1
			return records[len(records)-1], SeekSuccess
		}
	}

	// Branch block - navigate to first or last child
	keyLen := int(it.tag.KeyLen)
	entryLength := keyLen + 8
	offset := 12

	if first {
		// Get the first entry's child pointer
		if offset+entryLength > len(blockBuf) {
			return 0, SeekError
		}
		// Skip key and recNum
		offset += keyLen + 4
		// Read child pointer (big-endian)
		childPtr := int32(binary.BigEndian.Uint32(blockBuf[offset : offset+4]))
		return it.seekToExtreme(int64(childPtr), first)
	} else {
		// Get the last entry's child pointer
		lastEntryOffset := offset + (int(numKeys)-1)*entryLength
		if lastEntryOffset+entryLength > len(blockBuf) {
			return 0, SeekError
		}
		lastEntryOffset += keyLen + 4
		childPtr := int32(binary.BigEndian.Uint32(blockBuf[lastEntryOffset : lastEntryOffset+4]))
		return it.seekToExtreme(int64(childPtr), first)
	}
}

// Last positions at the last key in index order
// Returns the record number and true if successful, 0 and false if empty
func (it *CdxIterator) Last() (int32, bool) {
	if it.tag == nil || it.cdxFile == nil {
		it.eof = true
		it.bof = true
		return 0, false
	}

	recNo, result := it.seekToExtreme(int64(it.tag.RootBlock), false)
	if result == SeekError || result == SeekEOF {
		it.eof = true
		it.bof = true
		it.curRecNo = 0
		return 0, false
	}

	it.eof = false
	it.bof = false
	it.curRecNo = recNo
	return recNo, true
}

// EOF returns true if positioned past the last record
func (it *CdxIterator) EOF() bool {
	return it.eof
}

// BOF returns true if positioned before the first record
func (it *CdxIterator) BOF() bool {
	return it.bof
}

// RecNo returns the current record number
func (it *CdxIterator) RecNo() int32 {
	return it.curRecNo
}

// Next moves to the next record in index order
// Returns the record number and true if successful, 0 and false if at EOF
func (it *CdxIterator) Next() (int32, bool) {
	if it.eof || it.tag == nil || it.cdxFile == nil {
		return 0, false
	}

	// If we have cached records, try to advance within the current leaf
	if it.leafRecords != nil && it.curKeyIndex+1 < len(it.leafRecords) {
		it.curKeyIndex++
		it.curRecNo = it.leafRecords[it.curKeyIndex]
		it.bof = false
		return it.curRecNo, true
	}

	// Need to move to the next leaf block
	// Read current block to get rightNode pointer
	if it.curBlockOffset == 0 {
		// Not positioned - need to call First() first
		return 0, false
	}

	_, err := it.cdxFile.File.Seek(it.curBlockOffset, io.SeekStart)
	if err != nil {
		it.eof = true
		return 0, false
	}

	blockBuf := make([]byte, CDXBlockSize)
	n, err := it.cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		it.eof = true
		return 0, false
	}

	// Read rightNode pointer (offsets 8-11 in B4STD_HEADER)
	// This is a byte offset (4 bytes), not a block number
	rightNode := binary.LittleEndian.Uint32(blockBuf[8:12])
	if rightNode == 0 || rightNode == 0xFFFFFFFF {
		// No right sibling - we're at EOF
		it.eof = true
		return 0, false
	}

	// Navigate to the right sibling leaf block
	nextBlockOffset := int64(rightNode)
	_, err = it.cdxFile.File.Seek(nextBlockOffset, io.SeekStart)
	if err != nil {
		it.eof = true
		return 0, false
	}

	n, err = it.cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		it.eof = true
		return 0, false
	}

	// Parse the new leaf block
	numKeys := binary.LittleEndian.Uint16(blockBuf[2:4])
	if numKeys == 0 {
		it.eof = true
		return 0, false
	}

	records := extractCompactLeafRecords(blockBuf, int(numKeys))
	if len(records) == 0 {
		it.eof = true
		return 0, false
	}

	// Update iterator state
	it.curBlockOffset = nextBlockOffset
	it.leafRecords = records
	it.curKeyIndex = 0
	it.curRecNo = records[0]
	it.bof = false

	return it.curRecNo, true
}

// Previous moves to the previous record in index order
// Returns the record number and true if successful, 0 and false if at BOF
func (it *CdxIterator) Previous() (int32, bool) {
	if it.bof || it.tag == nil || it.cdxFile == nil {
		return 0, false
	}

	// If we have cached records, try to go back within the current leaf
	if it.leafRecords != nil && it.curKeyIndex > 0 {
		it.curKeyIndex--
		it.curRecNo = it.leafRecords[it.curKeyIndex]
		it.eof = false
		return it.curRecNo, true
	}

	// Need to move to the previous leaf block
	if it.curBlockOffset == 0 {
		return 0, false
	}

	_, err := it.cdxFile.File.Seek(it.curBlockOffset, io.SeekStart)
	if err != nil {
		it.bof = true
		return 0, false
	}

	blockBuf := make([]byte, CDXBlockSize)
	n, err := it.cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		it.bof = true
		return 0, false
	}

	// Read leftNode pointer (offsets 4-7 in B4STD_HEADER)
	// This is a byte offset (4 bytes), not a block number
	leftNode := binary.LittleEndian.Uint32(blockBuf[4:8])
	if leftNode == 0 || leftNode == 0xFFFFFFFF {
		// No left sibling - we're at BOF
		it.bof = true
		return 0, false
	}

	// Navigate to the left sibling leaf block
	prevBlockOffset := int64(leftNode)
	_, err = it.cdxFile.File.Seek(prevBlockOffset, io.SeekStart)
	if err != nil {
		it.bof = true
		return 0, false
	}

	n, err = it.cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		it.bof = true
		return 0, false
	}

	// Parse the new leaf block
	numKeys := binary.LittleEndian.Uint16(blockBuf[2:4])
	if numKeys == 0 {
		it.bof = true
		return 0, false
	}

	records := extractCompactLeafRecords(blockBuf, int(numKeys))
	if len(records) == 0 {
		it.bof = true
		return 0, false
	}

	// Update iterator state - position at LAST record in this leaf
	it.curBlockOffset = prevBlockOffset
	it.leafRecords = records
	it.curKeyIndex = len(records) - 1
	it.curRecNo = records[it.curKeyIndex]
	it.eof = false

	return it.curRecNo, true
}
