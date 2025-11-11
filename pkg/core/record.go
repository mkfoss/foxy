package core

type Record []byte

func (rec Record) Deleted() bool {

	return rec[0] == 0x2A
}
