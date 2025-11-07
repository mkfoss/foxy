package core

type HeaderDef struct {
	LastupdateYear  uint8
	LastupdateMonth uint8
	LastupdateDay   uint8
	RecordCount     uint32
	RecordOffset    uint16
	RecordSize      uint16
	reserved1       [16]byte
	TableFlags      uint8
	CodePage        uint8
	reserved2       [2]byte
}
