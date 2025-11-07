package mockfiler

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// MockFiler is a test implementation of Filer interface for testing
type MockFiler struct {
	Data      []byte
	Pos       int64
	Fileinfo  os.FileInfo
	FailSeek  error
	FailRead  error
	FailClose error
	FailStat  error
}

func NewMockFiler() *MockFiler {
	return &MockFiler{
		Data: make([]byte, 0),
	}
}

func (m *MockFiler) Stat() (os.FileInfo, error) {
	if m.FailStat != nil {
		return nil, m.FailStat
	}
	return m.Fileinfo, nil
}

func (m *MockFiler) Read(p []byte) (n int, err error) {

	if m.FailRead != nil {
		return 0, m.FailRead
	}
	if m.Pos >= int64(len(m.Data)) {
		return 0, io.EOF
	}
	n = copy(p, m.Data[m.Pos:])
	m.Pos += int64(n)
	return n, nil
}

func (m *MockFiler) Seek(offset int64, whence int) (int64, error) {

	if m.FailSeek != nil {
		return 0, m.FailSeek
	}
	switch whence {
	case io.SeekStart:
		m.Pos = offset
	case io.SeekCurrent:
		m.Pos += offset
	case io.SeekEnd:
		m.Pos = int64(len(m.Data)) + offset
	}
	return m.Pos, nil
}

func (m *MockFiler) Close() error {
	if m.FailClose != nil {
		return m.FailClose
	}
	return nil
}

func (m *MockFiler) Write(p any) error {

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	if err != nil {
		return err
	}

	m.Data = append(m.Data, buf.Bytes()...)
	return nil
}

func (m *MockFiler) MustWrite(p any) {
	err := m.Write(p)
	if err != nil {
		panic(err)
	}
}

func (m *MockFiler) ReadAt(p []byte, off int64) (n int, err error) {

	if off >= int64(len(m.Data)) {
		return 0, io.EOF
	}
	n = copy(p, m.Data[off:])
	return n, nil
}
