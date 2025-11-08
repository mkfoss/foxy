package core

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type Codepage uint8

func (cpi Codepage) String() string {
	switch cpi {
	case 0x03:
		return "Windows ANSI"
	default:
		return "Unknown codepage"
	}
}

func CodepageDecoder(codepage Codepage) *encoding.Decoder {
	switch codepage {
	case 0x03:
		return charmap.Windows1252.NewDecoder()
	default:
		return nil
	}
}

func CodepageSupported(codepage Codepage) bool {
	switch codepage {
	case 0x03:
		return true
	default:
		return false
	}
}

//	0x01: {0x01, 437, "U.S. MS-DOS"},
//	0x02: {0x02, 850, "International MS-DOS"},
//	0x03: {0x03, 1252, "Windows ANSI"},
//	0x04: {0x04, 10000, "Standard Macintosh"},
//	0x64: {0x64, 852, "Eastern European MS-DOS"},
//	0x65: {0x65, 866, "Russian MS-DOS"},
//	0x66: {0x66, 865, "Nordic MS-DOS"},
//	0x67: {0x67, 861, "Icelandic MS-DOS"},
//	0x68: {0x68, 895, "Kamenicky (Czech) MS-DOS"},
//	0x69: {0x69, 620, "Mazovia (Polish) MS-DOS"},
//	0x6A: {0x6A, 737, "Greek MS-DOS (437G)"},
//	0x6B: {0x6B, 857, "Turkish MS-DOS"},
//	0x78: {0x78, 950, "Traditional Chinese (Hong Kong SAR, Taiwan) Windows"},
//	0x79: {0x79, 949, "Korean Windows"},
//	0x7A: {0x7A, 936, "Chinese Simplified (PRC, Singapore) Windows"},
//	0x7B: {0x7B, 932, "Japanese Windows"},
//	0x7C: {0x7C, 874, "Thai Windows"},
//	0x7D: {0x7D, 1255, "Hebrew Windows"},
//	0x7E: {0x7E, 1256, "Arabic Windows"},
//	0xC8: {0xC8, 1250, "Eastern European Windows"},
//	0xC9: {0xC9, 1251, "Russian Windows"},
//	0xCA: {0xCA, 1254, "Turkish Windows"},
//	0xCB: {0xCB, 1253, "Greek Windows"},
//	0x96: {0x96, 10007, "Russian Macintosh"},
//	0x97: {0x97, 10029, "Macintosh EE"},
//	0x98: {0x98, 10006, "Greek Macintosh"},
