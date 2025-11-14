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
	tagNames := make(map[string]int64)

	// Seek to block 0 (compound tag header)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Read compound tag header
	headerBuf := make([]byte, CDXBlockSize)
	n, err := file.Read(headerBuf)
	if err != nil || n != CDXBlockSize {
		return nil, err
	}

	rootBlock := binary.LittleEndian.Uint32(headerBuf[0:4])
	keyLen := binary.LittleEndian.Uint16(headerBuf[12:14])

	// Navigate to root block
	if _, err := file.Seek(int64(rootBlock), io.SeekStart); err != nil {
		return nil, err
	}

	rootBuf := make([]byte, CDXBlockSize)
	n, err = file.Read(rootBuf)
	if err != nil || n != CDXBlockSize {
		return nil, err
	}

	blockType := binary.LittleEndian.Uint16(rootBuf[0:2])
	numKeys := binary.LittleEndian.Uint16(rootBuf[2:4])

	// If branch, navigate to first leaf
	if (blockType & 0x03) <= 1 {
		// Follow first child pointer
		offset := 12 + int(keyLen) + 4
		if offset+4 > len(rootBuf) {
			return nil, fmt.Errorf("invalid compound tag structure")
		}
		childPtr := binary.BigEndian.Uint32(rootBuf[offset : offset+4])

		if _, err := file.Seek(int64(childPtr), io.SeekStart); err != nil {
			return nil, err
		}

		n, err = file.Read(rootBuf)
		if err != nil || n != CDXBlockSize {
			return nil, err
		}

		numKeys = binary.LittleEndian.Uint16(rootBuf[2:4])
	}

	// Parse compact leaf to extract tag names
	recNumMask := binary.LittleEndian.Uint32(rootBuf[14:18])
	shortBytes := int(rootBuf[23])
	dupCntBits := rootBuf[21]
	trailCntBits := rootBuf[22]
	dupCntMask := rootBuf[18]
	trailCntMask := rootBuf[19]

	if shortBytes == 0 || shortBytes > 6 {
		return nil, fmt.Errorf("invalid shortBytes value: %d", shortBytes)
	}

	infoStart := 24
	extSpace := CDXBlockSize - 24
	kPos := extSpace

	keyData := make([]byte, keyLen)
	for j := range keyData {
		keyData[j] = ' '
	}

	for i := 0; i < int(numKeys); i++ {
		v := i * shortBytes

		if infoStart+v+shortBytes > len(rootBuf) {
			break
		}

		// Read control word
		c := binary.LittleEndian.Uint16(rootBuf[infoStart+v+shortBytes-2:])
		trailCnt := int((c >> (16 - trailCntBits)) & uint16(trailCntMask))
		dupCnt := int((c >> (16 - (trailCntBits + dupCntBits))) & uint16(dupCntMask))

		// Read record number (tag header offset)
		var rawRecNo uint32
		switch shortBytes {
		case 2:
			rawRecNo = uint32(binary.LittleEndian.Uint16(rootBuf[infoStart+v:]))
		case 3:
			rawRecNo = uint32(rootBuf[infoStart+v]) |
				(uint32(rootBuf[infoStart+v+1]) << 8) |
				(uint32(rootBuf[infoStart+v+2]) << 16)
		case 4:
			rawRecNo = binary.LittleEndian.Uint32(rootBuf[infoStart+v:])
		}
		tagOffset := int64(rawRecNo & recNumMask)

		// Reconstruct key (tag name)
		kPos = kPos - int(keyLen) + dupCnt + trailCnt
		newBytes := int(keyLen) - dupCnt - trailCnt
		if newBytes > 0 {
			if infoStart+kPos+newBytes > len(rootBuf) {
				break
			}
			copy(keyData[dupCnt:dupCnt+newBytes], rootBuf[infoStart+kPos:infoStart+kPos+newBytes])
		}
		for j := int(keyLen) - trailCnt; j < int(keyLen); j++ {
			keyData[j] = 0
		}

		// Trim nulls and spaces to get tag name
		tagName := string(keyData)
		for len(tagName) > 0 && (tagName[len(tagName)-1] == 0 || tagName[len(tagName)-1] == ' ') {
			tagName = tagName[:len(tagName)-1]
		}

		tagNames[tagName] = tagOffset
	}

	return tagNames, nil
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
	keyLen := binary.LittleEndian.Uint16(headerBuf[12:14])
	typeCode := headerBuf[14]
	signature := headerBuf[15]

	// Verify signature
	if signature != CDXSignature {
		return nil, fmt.Errorf("invalid tag header signature")
	}

	// Determine if descending (bit 0x40 in typeCode)
	descending := (typeCode & 0x40) != 0

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
		TypeCode:   typeCode,
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
