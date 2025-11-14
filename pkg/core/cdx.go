package core

import (
	"encoding/binary"
	"fmt"
	"io"
)

// CDX constants
const (
	CDXBlockSize    = 512
	CDXHeaderSize   = 512
	CDXSignature    = 0x01
	CDXTagNameSize  = 11
	CDXCompoundType = 0x7C
)

// SeekResult represents the result of a CDX seek operation
type SeekResult int

const (
	SeekSuccess SeekResult = iota // Found exact match
	SeekAfter                     // Positioned after the seek value
	SeekEOF                       // Past end of index
	SeekError                     // Error occurred
)

// CdxFile represents an open CDX index file
type CdxFile struct {
	File io.ReadSeeker
	Tags []*CdxTag
}

// CdxTag represents a single index tag within a CDX file
type CdxTag struct {
	Name       string
	Expression string
	RootBlock  uint32 // Byte offset to root block
	KeyLen     uint16
	TypeCode   byte
	Descending bool
	HeaderPos  int64
}

// CdxHeader represents the header structure of a CDX file or tag
type CdxHeader struct {
	RootBlock  uint32
	FreeBlock  uint32
	Version    byte
	KeyLen     uint16
	Options    byte
	Signature  byte
	KeyExpr    string
	ForExpr    string
}

// OpenCdx opens a CDX index file and parses its tag metadata
func OpenCdx(file io.ReadSeeker) (*CdxFile, error) {
	if file == nil {
		return nil, fmt.Errorf("nil file provided")
	}

	cdxFile := &CdxFile{
		File: file,
		Tags: make([]*CdxTag, 0),
	}

	// Parse tags from compound tag (block 0)
	if err := parseCdxTags(cdxFile); err != nil {
		return nil, fmt.Errorf("failed to parse CDX tags: %w", err)
	}

	return cdxFile, nil
}

// parseCdxTags reads the compound tag at block 0 to find all tag headers
func parseCdxTags(cdxFile *CdxFile) error {
	// Read compound tag at offset 0
	compoundData, err := readCompoundTag(cdxFile.File)
	if err != nil {
		return fmt.Errorf("failed to read compound tag: %w", err)
	}

	// Parse tag entries from compound data
	// Each entry maps a tag name to its header offset
	for tagName, headerOffset := range compoundData {
		tag, err := readTagHeader(cdxFile.File, headerOffset, tagName)
		if err != nil {
			// Skip invalid tags
			continue
		}
		cdxFile.Tags = append(cdxFile.Tags, tag)
	}

	return nil
}

// readCompoundTag reads the compound tag at block 0
// Returns a map of tag names to their header offsets
func readCompoundTag(file io.ReadSeeker) (map[string]int64, error) {
	// Seek to block 0
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Read header block
	headerBuf := make([]byte, CDXHeaderSize)
	if _, err := io.ReadFull(file, headerBuf); err != nil {
		return nil, err
	}

	// Verify signature
	signature := headerBuf[8]
	if signature != CDXSignature {
		return nil, fmt.Errorf("invalid CDX signature: expected 0x%02X, got 0x%02X", CDXSignature, signature)
	}

	// Read root block offset (little-endian uint32)
	rootBlock := binary.LittleEndian.Uint32(headerBuf[0:4])

	// Read the compound tag block
	if _, err := file.Seek(int64(rootBlock), io.SeekStart); err != nil {
		return nil, err
	}

	blockBuf := make([]byte, CDXBlockSize)
	if _, err := io.ReadFull(file, blockBuf); err != nil {
		return nil, err
	}

	// Parse block header
	blockType := binary.LittleEndian.Uint16(blockBuf[0:2])
	numKeys := binary.LittleEndian.Uint16(blockBuf[2:4])

	// Verify this is a compound tag block
	if blockType != CDXCompoundType {
		return nil, fmt.Errorf("expected compound tag block type 0x%02X, got 0x%02X", CDXCompoundType, blockType)
	}

	// Parse tag entries
	tags := make(map[string]int64)
	offset := 12 // Start after block header

	for i := 0; i < int(numKeys); i++ {
		// Each entry is: tag name (11 bytes) + header offset (4 bytes big-endian)
		if offset+CDXTagNameSize+4 > len(blockBuf) {
			break
		}

		// Extract tag name (null-terminated or space-padded)
		nameBytes := blockBuf[offset : offset+CDXTagNameSize]
		tagName := ""
		for _, b := range nameBytes {
			if b == 0 || b == ' ' {
				break
			}
			tagName += string(b)
		}
		offset += CDXTagNameSize

		// Extract header offset (big-endian)
		headerOffset := int64(binary.BigEndian.Uint32(blockBuf[offset : offset+4]))
		offset += 4

		if tagName != "" {
			tags[tagName] = headerOffset
		}
	}

	return tags, nil
}

// readTagHeader reads a tag header at the specified offset
func readTagHeader(file io.ReadSeeker, offset int64, tagName string) (*CdxTag, error) {
	// Seek to header position
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	// Read header block
	headerBuf := make([]byte, CDXHeaderSize)
	if _, err := io.ReadFull(file, headerBuf); err != nil {
		return nil, err
	}

	// Parse header fields (little-endian)
	rootBlock := binary.LittleEndian.Uint32(headerBuf[0:4])
	// freeBlock := binary.LittleEndian.Uint32(headerBuf[4:8])
	// version := headerBuf[8]
	keyLen := binary.LittleEndian.Uint16(headerBuf[10:12])
	options := headerBuf[12]
	signature := headerBuf[13]

	// Verify signature
	if signature != CDXSignature {
		return nil, fmt.Errorf("invalid tag header signature")
	}

	// Determine if descending
	descending := (options & 0x08) != 0

	// Extract key expression (starts at offset 528, null-terminated)
	keyExprStart := 528
	keyExpr := ""
	for i := keyExprStart; i < len(headerBuf); i++ {
		if headerBuf[i] == 0 {
			break
		}
		keyExpr += string(headerBuf[i])
	}

	tag := &CdxTag{
		Name:       tagName,
		Expression: keyExpr,
		RootBlock:  rootBlock,
		KeyLen:     keyLen,
		TypeCode:   headerBuf[14], // Type code at offset 14
		Descending: descending,
		HeaderPos:  offset,
	}

	return tag, nil
}

// FindTag searches for a tag by name (case-insensitive)
func (cf *CdxFile) FindTag(tagName string) *CdxTag {
	for _, tag := range cf.Tags {
		if equalIgnoreCase(tag.Name, tagName) {
			return tag
		}
	}
	return nil
}

// equalIgnoreCase performs case-insensitive string comparison
func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		// Convert to uppercase for comparison
		if ca >= 'a' && ca <= 'z' {
			ca -= 32
		}
		if cb >= 'a' && cb <= 'z' {
			cb -= 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
