package core

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

// Seek searches for a key in the specified tag
// Returns the record number and seek result
func (tag *CdxTag) Seek(cdxFile *CdxFile, searchKey string) (int32, SeekResult) {
	if cdxFile == nil || cdxFile.File == nil {
		return 0, SeekError
	}

	// Pad or truncate search key to match key length
	paddedKey := padKey(searchKey, int(tag.KeyLen))

	// Start from root block (RootBlock is a byte offset)
	recNo, result := tag.seekInBlock(cdxFile, int64(tag.RootBlock), paddedKey, 0)

	return recNo, result
}

// seekInBlock recursively searches through B+ tree blocks
func (tag *CdxTag) seekInBlock(cdxFile *CdxFile, blockOffset int64, searchKey string, depth int) (int32, SeekResult) {
	// Safety: prevent infinite recursion
	if depth > 20 {
		return 0, SeekError
	}

	// Read the block
	_, err := cdxFile.File.Seek(blockOffset, io.SeekStart)
	if err != nil {
		return 0, SeekError
	}

	blockBuf := make([]byte, CDXBlockSize)
	n, err := cdxFile.File.Read(blockBuf)
	if err != nil || n != CDXBlockSize {
		return 0, SeekError
	}

	// Parse block header
	blockType := binary.LittleEndian.Uint16(blockBuf[0:2])
	numKeys := binary.LittleEndian.Uint16(blockBuf[2:4])

	if numKeys == 0 {
		return 0, SeekEOF
	}

	// Determine if this is a leaf block
	// (blockType & 0x03) > 1 means leaf (2 or 3), 0 or 1 means branch
	isLeaf := (blockType & 0x03) > 1

	if isLeaf {
		// Leaf block - VFP uses compact format for all leaf blocks
		return parseCompactLeafBlock(blockBuf, int(tag.KeyLen), searchKey, int(numKeys))
	}

	// Branch block - find child pointer to follow
	childBlock := tag.searchBranchBlock(blockBuf, searchKey, int(numKeys))
	if childBlock == 0 {
		return 0, SeekEOF
	}

	// Recursively search child block (child pointers are byte offsets)
	return tag.seekInBlock(cdxFile, int64(childBlock), searchKey, depth+1)
}

// searchBranchBlock searches for the child pointer to follow in a branch block
func (tag *CdxTag) searchBranchBlock(blockBuf []byte, searchKey string, numKeys int) int32 {
	keyLen := int(tag.KeyLen)
	// EntryLength = KeyLength + 8 (4 bytes recNum + 4 bytes childPtr)
	entryLength := keyLen + 8
	offset := 12 // Start after header (Node_Atr, Entry_Ct, Left_Ptr, Rght_Ptr)

	// In VFP CDX branch blocks:
	// Each entry is: [Key (keyLen bytes)][RecNum (4 bytes, big-endian)][ChildPtr (4 bytes, big-endian)]
	var leftChild int32

	// Search through keys
	for i := 0; i < numKeys; i++ {
		if offset+entryLength > len(blockBuf) {
			break
		}

		// Extract key (keyLen bytes)
		keyData := blockBuf[offset : offset+keyLen]
		offset += keyLen

		// Skip record number (4 bytes, big-endian)
		offset += 4

		// Extract child block pointer (4 bytes, big-endian)
		childPtr := int32(binary.BigEndian.Uint32(blockBuf[offset : offset+4]))
		offset += 4

		// Compare keys
		cmp := compareKeys([]byte(searchKey), keyData)
		if cmp <= 0 {
			// searchKey <= current key, follow this entry's child
			return childPtr
		}

		// searchKey > current key, remember this child and continue
		leftChild = childPtr
	}

	// searchKey is greater than all keys, follow last child pointer
	return leftChild
}

// compareKeys compares two key byte slices
// Returns: -1 if key1 < key2, 0 if equal, 1 if key1 > key2
func compareKeys(key1, key2 []byte) int {
	// Trim trailing spaces for comparison
	k1 := bytes.TrimRight(key1, " ")
	k2 := bytes.TrimRight(key2, " ")

	return bytes.Compare(k1, k2)
}

// padKey pads or truncates a key to the specified length
func padKey(key string, length int) string {
	if len(key) >= length {
		return key[:length]
	}
	return key + strings.Repeat(" ", length-len(key))
}
