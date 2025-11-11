package foxy

type Record struct {
	data []byte
}

func (rec *Record) Deleted() bool {

	return rec.data[0] == 0x2A
}
