package foxy

import (
	"fmt"
	"io"
)

type DefaulNavigator struct {
	offset   int
	size     int
	count    int
	readfunc ReadFunc
	rdr      io.ReadSeeker
	position int
}

func (defnav *DefaulNavigator) IsFirst() bool {
	return defnav.position == 1
}

func (defnav *DefaulNavigator) IsLast() bool {
	return defnav.position == defnav.count
}

func (defnav *DefaulNavigator) First() error {
	if _, err := defnav.rdr.Seek(int64(defnav.offset), io.SeekStart); err != nil {
		return NewNavigationError().SetContext("seek first").SetWrapped(err)
	}
	defnav.position = 1
	return defnav.readfunc()
}

func (defnav *DefaulNavigator) Next() error {
	if defnav.position == (defnav.count) {
		return NewNavigationEofError().SetContext("seek next")
	}
	defnav.position++
	return defnav.readfunc()
}

func (defnav *DefaulNavigator) Previous() error {
	if defnav.position == 1 {
		return NewNavigationBofError().SetContext("seek previous")
	}
	_, err := defnav.rdr.Seek(int64(-defnav.size*2), io.SeekCurrent)
	if err != nil {
		return NewNavigationError().SetContext("seek previous").SetWrapped(err)
	}
	defnav.position--
	return defnav.readfunc()
}

func (defnav *DefaulNavigator) Last() error {
	if _, err := defnav.rdr.Seek(int64(defnav.offset+(defnav.size*defnav.count)-(defnav.size)), io.SeekStart); err != nil {
		return NewNavigationError().SetContext("seek last").SetWrapped(err)
	}
	defnav.position = defnav.count
	return defnav.readfunc()
}

func (defnav *DefaulNavigator) Goto(position int) error {
	if position < 0 || position >= defnav.count {
		return NewNavigationError().SetContext(fmt.Sprintf("goto %d", defnav.position))
	}

	pos := int64(defnav.offset + (defnav.size * (position - 1)))
	if _, err := defnav.rdr.Seek(pos, io.SeekStart); err != nil {
		return NewNavigationError()
	}

	defnav.position = position
	return defnav.readfunc()
}

func (defnav *DefaulNavigator) Position() int {
	return defnav.position
}

func (defnav *DefaulNavigator) Initialize(readseeker io.ReadSeeker, offset, size, count int, rf ReadFunc) error {
	defnav.rdr = readseeker
	defnav.offset = offset
	defnav.size = size
	defnav.count = count
	defnav.readfunc = rf

	return nil
}

func (defnav *DefaulNavigator) Finalize() error {
	return nil
}
