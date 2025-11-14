package core

import (
	"bytes"
	"encoding/binary"
)

// parseCompactLeafBlock parses a VFP compact format leaf block.
// This implementation is based on the CodeBase C implementation
// in b4block.c (b4key, x4recNo, x4dupCnt, x4trailCnt functions).
//
// VFP compact format structure:
// - B4STD_HEADER (12 bytes): nodeAttribute, nKeys, leftNode, rightNode
// - B4NODE_HEADER (12 bytes): freeSpace, recNumMask, various bit lengths
// - Info table: infoLen bytes per key (typically 2-3 bytes)
// - Free space
// - Key data pool: keys stored backwards from end of block with compression
//
// Each info word contains packed bit fields:
// - recNumLen bits: record number
// - dupCntLen bits: duplicate byte count
// - trailCntLen bits: trailing byte count
func parseCompactLeafBlock(blockBuf []byte, keyLen int, searchKey string, numKeys int) (int32, SeekResult) {
	if len(blockBuf) < 24 {
		return 0, SeekError
	}

	if numKeys == 0 {
		return 0, SeekEOF
	}

	// Read metadata from block (B4NODE_HEADER at offset 12-23)
	// The B4NODE_HEADER structure:
	// Offset 12-13: FreeSpace (Word)
	// Offset 14-17: RecNumMask (LongInt)
	// Offset 18: DupCntMask (Byte)
	// Offset 19: TrlCntMask (Byte)
	// Offset 20: RecNumBits (Byte)
	// Offset 21: DupCntBits (Byte)
	// Offset 22: TrlCntBits (Byte)
	// Offset 23: ShortBytes (Byte)
	recNumMask := binary.LittleEndian.Uint32(blockBuf[14:18])
	dupCntMask := blockBuf[18]
	trailCntMask := blockBuf[19]
	dupCntBits := blockBuf[21]
	trailCntBits := blockBuf[22]
	shortBytes := int(blockBuf[23])

	if shortBytes == 0 || shortBytes > 6 {
		return 0, SeekError
	}

	// Info table starts at offset 24 (12 byte B4STD_HEADER + 12 byte B4NODE_HEADER)
	infoTableStart := 24

	// Calculate extSpace - free space available for keys
	extSpace := CDXBlockSize - 24

	// Key data pool position starts at extSpace and works backwards
	kPos := extSpace

	// Build key buffer (reused for all keys)
	keyData := make([]byte, keyLen)
	for j := range keyData {
		keyData[j] = ' '
	}

	searchKeyBytes := []byte(searchKey)

	// Process each key sequentially
	for i := 0; i < numKeys; i++ {
		// Position in info table
		v := i * shortBytes

		// Read the last 2 bytes of the info word for trail/dup counts
		if infoTableStart+v+shortBytes > len(blockBuf) {
			return 0, SeekError
		}
		c := binary.LittleEndian.Uint16(blockBuf[infoTableStart+v+shortBytes-2 : infoTableStart+v+shortBytes])

		// Extract trail count
		trailCnt := int((c >> (16 - trailCntBits)) & uint16(trailCntMask))

		// Extract dup count
		dupCnt := int((c >> (16 - (trailCntBits + dupCntBits))) & uint16(dupCntMask))

		// Extract record number
		var recNo uint32
		if infoTableStart+v+shortBytes <= len(blockBuf) {
			var rawRecNo uint32
			switch shortBytes {
			case 2:
				rawRecNo = uint32(binary.LittleEndian.Uint16(blockBuf[infoTableStart+v:]))
			case 3:
				rawRecNo = uint32(blockBuf[infoTableStart+v]) |
					(uint32(blockBuf[infoTableStart+v+1]) << 8) |
					(uint32(blockBuf[infoTableStart+v+2]) << 16)
			default:
				if infoTableStart+v+4 <= len(blockBuf) {
					rawRecNo = binary.LittleEndian.Uint32(blockBuf[infoTableStart+v:])
				} else {
					return 0, SeekError
				}
			}
			recNo = rawRecNo & recNumMask
		} else {
			return 0, SeekError
		}

		// Update key position
		kPos = kPos - keyLen + dupCnt + trailCnt

		// Copy new key bytes
		newBytes := keyLen - dupCnt - trailCnt
		if newBytes > 0 {
			if kPos < 0 || kPos+newBytes > len(blockBuf) {
				return 0, SeekError
			}
			copy(keyData[dupCnt:dupCnt+newBytes], blockBuf[infoTableStart+kPos:infoTableStart+kPos+newBytes])
		}

		// Fill trailing bytes
		for j := keyLen - trailCnt; j < keyLen; j++ {
			keyData[j] = 0
		}

		// Compare with search key
		cmp := bytes.Compare(bytes.TrimRight(searchKeyBytes, " "), bytes.TrimRight(keyData, " \x00"))
		if cmp == 0 {
			return int32(recNo), SeekSuccess
		} else if cmp < 0 {
			return int32(recNo), SeekAfter
		}
	}

	return 0, SeekEOF
}

// extractCompactLeafRecords extracts all record numbers from a compact leaf block
// Returns the records in order from first to last
func extractCompactLeafRecords(blockBuf []byte, keyLen int, numKeys int) ([]int32, error) {
	if len(blockBuf) < 24 || numKeys == 0 {
		return nil, nil
	}

	// Read metadata from block (B4NODE_HEADER at offset 12-23)
	recNumMask := binary.LittleEndian.Uint32(blockBuf[14:18])
	shortBytes := int(blockBuf[23])

	if shortBytes == 0 || shortBytes > 6 {
		return nil, nil
	}

	infoTableStart := 24
	records := make([]int32, 0, numKeys)

	// Extract record number from each entry
	for i := 0; i < numKeys; i++ {
		v := i * shortBytes
		if infoTableStart+v+shortBytes > len(blockBuf) {
			break
		}

		// Read shortBytes and mask
		var rawRecNo uint32
		switch shortBytes {
		case 2:
			rawRecNo = uint32(binary.LittleEndian.Uint16(blockBuf[infoTableStart+v:]))
		case 3:
			rawRecNo = uint32(blockBuf[infoTableStart+v]) |
				(uint32(blockBuf[infoTableStart+v+1]) << 8) |
				(uint32(blockBuf[infoTableStart+v+2]) << 16)
		default:
			if infoTableStart+v+4 <= len(blockBuf) {
				rawRecNo = binary.LittleEndian.Uint32(blockBuf[infoTableStart+v:])
			} else {
				break
			}
		}

		recNo := rawRecNo & recNumMask
		records = append(records, int32(recNo))
	}

	return records, nil
}
