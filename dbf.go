package foxy

import (
	"io"
	"os"
	"time"

	"github.com/mkfoss/foxy/pkg/core"
)

type Dbf struct {
	filename     string
	dbftype      core.Dbftype
	opener       Opener
	fl           io.ReadSeekCloser
	lastupdated  time.Time
	recordoffset int
	recordcount  int
	recordsize   int
	hasindex     bool
	hasfpt       bool
	codepage     core.Codepage
}

func (dbf *Dbf) Open(name string) error {

	return dbf.OpenWithOpener(name, &OsOpener{})
}

func (dbf *Dbf) OpenWithOpener(name string, opener Opener) error {

	fl, err := opener.OpenFile(name, os.O_RDONLY, 0600)
	if err != nil {
		return NewErrorf("open dbf %s failed", name).SetContext("open with opener").SetWrapped(err)
	}
	dbft, err := core.ReadDbftype(fl)
	if err != nil {
		return NewError("read magic byte failed").SetContext("open with opener").SetWrapped(err)
	}

	dbf.dbftype = dbft

	hdrdef, err := core.ReadDbfHeaderDef(fl)
	if err != nil {
		return NewError("read header failed").SetContext("open with opener").SetWrapped(err)
	}

	luyr := int(hdrdef.LastupdateYear) + 2000
	if hdrdef.LastupdateYear >= 70 {
		luyr -= 100
	}

	dbf.lastupdated = time.Date(luyr, time.Month(hdrdef.LastupdateMonth), int(hdrdef.LastupdateDay), 0, 0, 0, 0, time.UTC)
	dbf.recordoffset = int(hdrdef.RecordOffset)
	dbf.recordsize = int(hdrdef.RecordSize)
	dbf.recordcount = int(hdrdef.RecordCount)
	dbf.hasindex = hdrdef.TableFlags&0x01 == 0x01
	dbf.hasfpt = hdrdef.TableFlags&0x02 == 0x02
	dbf.codepage = hdrdef.CodePage

	//todo: sanity checking

	//these should be set after everything is initialized
	dbf.opener = opener
	dbf.fl = fl
	dbf.filename = name

	return nil
}

func (dbf *Dbf) Close() error {

	err := dbf.fl.Close()
	if err != nil {
		return NewErrorf("close dbf %s failed", dbf.filename).SetWrapped(err).SetContext("close dbf")
	}
	dbf.fl = nil
	dbf.opener = nil
	dbf.filename = ""

	dbf.lastupdated = time.Time{}
	dbf.recordoffset = 0
	dbf.recordsize = 0
	dbf.recordcount = 0
	dbf.hasindex = false
	dbf.hasfpt = false
	dbf.codepage = 0

	return nil
}

func (dbf *Dbf) LastUpdated() time.Time {
	return dbf.lastupdated
}

func (dbf *Dbf) RecordOffset() int {
	return dbf.recordoffset
}

func (dbf *Dbf) RecordCount() int {
	return dbf.recordcount
}

func (dbf *Dbf) RecordSize() int {
	return dbf.recordsize
}

func (dbf *Dbf) HasIndex() bool {
	return dbf.hasindex
}

func (dbf *Dbf) HasFpt() bool {
	return dbf.hasfpt
}

func (dbf *Dbf) CodePage() core.Codepage {
	return dbf.codepage
}

func (dbf *Dbf) Active() bool {
	return dbf.fl != nil
}

func (dbf *Dbf) Filename() string {
	return dbf.filename
}
