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
	rec          []byte

	*Fields
	*core.Record
	Navigator
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

	dbf.opener = opener
	dbf.fl = fl
	dbf.filename = name

	flds := &Fields{}
	if err = flds.Read(dbf); err != nil {
		return NewError("read fields failed").SetContext("open with opener").SetWrapped(err)
	}
	dbf.Fields = flds

	dbf.Record = core.NewRecord(dbf.recordsize, dbf.codepage)

	err = dbf.SetNavigator(&DefaulNavigator{})
	if err != nil {
		return NewError("set navigator failed").SetContext("open with opener").SetWrapped(err)
	}

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

	dbf.Fields.fields = make([]*Field, 0)
	dbf.Fields.fieldmap = make(map[string]int)
	dbf.Fields = nil

	dbf.Record.ClearData()
	dbf.Record = nil

	err = dbf.Navigator.Finalize()
	if err != nil {
		return NewNavigationError().SetContext("dbf close, finalize navigator").SetWrapped(err)
	}
	dbf.Navigator = nil

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

func (dbf *Dbf) SetNavigator(navi Navigator) error {
	if !dbf.Active() {
		return NewInactiveError().SetContext("set navigator")
	}

	if dbf.Navigator != nil {
		if err := dbf.Navigator.Finalize(); err != nil {
			return NewError("navigator finalize failed").SetContext("set navigator").SetWrapped(err)
		}
	}

	dbf.Navigator = navi
	if err := dbf.Navigator.Initialize(dbf.fl, dbf.recordoffset, dbf.recordsize, dbf.recordcount, dbf.readFunc); err != nil {
		return NewError("navigator initialize failed").SetContext("set navigator").SetWrapped(err)
	}

	if err := dbf.Navigator.First(); err != nil {
		return NewNavigationError().SetContext("set navigator, first").SetWrapped(err)
	}

	return nil
}

func (dbf *Dbf) readFunc() error {
	//assumed that reader offset is correct
	err := dbf.Record.LoadData(dbf.fl)
	if err != nil {
		return NewErrorf("read record data failed").SetContext("read func").SetWrapped(err)
	}

	return nil
}
