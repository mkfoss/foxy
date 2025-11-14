package foxy

import (
	"encoding/binary"
	"io"
	"os"
	"time"

	"github.com/mkfoss/foxy/pkg/core"
)

type Dbf struct {
	filename           string
	dbftype            core.Dbftype
	opener             Opener
	fl                 io.ReadSeekCloser
	lastupdated        time.Time
	recordoffset       int
	recordcount        int
	recordsize         int
	hasindex           bool
	hasfpt             bool
	codepage           core.Codepage
	rec                []byte
	fpt                io.ReadSeekCloser
	fptFilename        string
	fptBlockSize       uint16
	cdx                *core.CdxFile
	cdxFilename        string
	UseFuzzyFileSearch bool

	*Fields
	*core.Record
	Navigator
}
func (dbf *Dbf) Open(name string) error {

	return dbf.OpenWithOpener(name, &OsOpener{})
}

func (dbf *Dbf) OpenWithOpener(name string, opener Opener) error {
	// Use fuzzy file search if enabled and opener supports listing
	dbfFilename := name
	if dbf.UseFuzzyFileSearch {
		// Check if opener supports OpenerLister interface
		if lister, ok := opener.(core.OpenerLister); ok {
			foundFile, err := core.FuzzyFindFile(name, core.FTDbf, lister)
			if err != nil {
				return NewErrorf("fuzzy search for dbf %s failed", name).SetContext("open with opener").SetWrapped(err)
			}
			dbfFilename = foundFile
		} else {
			return NewError("fuzzy file search enabled but opener does not support OpenerLister interface").SetContext("open with opener")
		}
	}

	fl, err := opener.OpenFile(dbfFilename, os.O_RDONLY, 0600)
	if err != nil {
		return NewErrorf("open dbf %s failed", dbfFilename).SetContext("open with opener").SetWrapped(err)
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
	dbf.filename = dbfFilename

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

	// Close FPT file if open
	if dbf.fpt != nil {
		if err := dbf.fpt.Close(); err != nil {
			return NewErrorf("close fpt %s failed", dbf.fptFilename).SetWrapped(err).SetContext("close dbf")
		}
		dbf.fpt = nil
		dbf.fptFilename = ""
		dbf.fptBlockSize = 0
	}

	// Close CDX file if open
	if dbf.cdx != nil {
		if closer, ok := dbf.cdx.File.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return NewErrorf("close cdx %s failed", dbf.cdxFilename).SetWrapped(err).SetContext("close dbf")
			}
		}
		dbf.cdx = nil
		dbf.cdxFilename = ""
	}

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

func (dbf *Dbf) FptFilename() string {
	return dbf.fptFilename
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

	// If this is a CdxNavigator, set the Dbf reference
	if cdxNav, ok := navi.(*CdxNavigator); ok {
		cdxNav.SetDbf(dbf)
	}

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

func (dbf *Dbf) ensureFptLoaded() error {
	// Already loaded
	if dbf.fpt != nil {
		return nil
	}

	// Check if FPT exists according to header
	if !dbf.hasfpt {
		return NewError("dbf header indicates no fpt file")
	}

	// Determine FPT filename base (strip extension from DBF filename)
	baseFilename := dbf.filename
	dbfExt := ""
	for i := len(dbf.filename) - 1; i >= 0; i-- {
		if dbf.filename[i] == '.' {
			dbfExt = dbf.filename[i:]
			baseFilename = dbf.filename[:i]
			break
		}
	}

	// Use fuzzy file search if enabled and opener supports listing, otherwise construct the filename
	var fptFilename string
	if dbf.UseFuzzyFileSearch {
		// Check if opener supports OpenerLister interface
		if lister, ok := dbf.opener.(core.OpenerLister); ok {
			foundFile, err := core.FuzzyFindFile(baseFilename, core.FTFpt, lister)
			if err != nil {
				return NewErrorf("fuzzy search for fpt file failed").SetWrapped(err).SetContext("ensure fpt loaded")
			}
			fptFilename = foundFile
		} else {
			return NewError("fuzzy file search enabled but opener does not support OpenerLister interface").SetContext("ensure fpt loaded")
		}
	} else {
		// Match case of DBF extension
		fptExt := ".fpt"
		if len(dbfExt) > 0 && dbfExt[0] == '.' {
			if dbfExt[1:2] == "D" || dbfExt[1:2] == "d" {
				if dbfExt[1:2] == "D" {
					fptExt = ".FPT"
				}
			}
		}
		fptFilename = baseFilename + fptExt
	}

	// Open FPT file
	fptFile, err := dbf.opener.OpenFile(fptFilename, os.O_RDONLY, 0600)
	if err != nil {
		return NewErrorf("failed to open FPT file %s", fptFilename).SetWrapped(err).SetContext("ensure fpt loaded")
	}

	// Read FPT header to get block size
	// FPT header: 4 bytes next free block, 2 bytes unused, 2 bytes block size (all big-endian)
	var nextFree uint32
	var unused [2]byte
	var blockSize uint16

	if err := binary.Read(fptFile, binary.BigEndian, &nextFree); err != nil {
		_ = fptFile.Close()
		return NewError("failed to read FPT header").SetWrapped(err).SetContext("ensure fpt loaded")
	}
	if err := binary.Read(fptFile, binary.BigEndian, &unused); err != nil {
		_ = fptFile.Close()
		return NewError("failed to read FPT header").SetWrapped(err).SetContext("ensure fpt loaded")
	}
	if err := binary.Read(fptFile, binary.BigEndian, &blockSize); err != nil {
		_ = fptFile.Close()
		return NewError("failed to read FPT header").SetWrapped(err).SetContext("ensure fpt loaded")
	}

	dbf.fpt = fptFile
	dbf.fptFilename = fptFilename
	dbf.fptBlockSize = blockSize

	return nil
}

func (dbf *Dbf) ensureCdxLoaded() error {
	// Already loaded
	if dbf.cdx != nil {
		return nil
	}

	// Check if CDX exists according to header
	if !dbf.hasindex {
		return NewError("dbf header indicates no cdx file")
	}

	// Determine CDX filename base (strip extension from DBF filename)
	baseFilename := dbf.filename
	dbfExt := ""
	for i := len(dbf.filename) - 1; i >= 0; i-- {
		if dbf.filename[i] == '.' {
			dbfExt = dbf.filename[i:]
			baseFilename = dbf.filename[:i]
			break
		}
	}

	// Use fuzzy file search if enabled and opener supports listing, otherwise construct the filename
	var cdxFilename string
	if dbf.UseFuzzyFileSearch {
		// Check if opener supports OpenerLister interface
		if lister, ok := dbf.opener.(core.OpenerLister); ok {
			foundFile, err := core.FuzzyFindFile(baseFilename, core.FTCdx, lister)
			if err != nil {
				return NewErrorf("fuzzy search for cdx file failed").SetWrapped(err).SetContext("ensure cdx loaded")
			}
			cdxFilename = foundFile
		} else {
			return NewError("fuzzy file search enabled but opener does not support OpenerLister interface").SetContext("ensure cdx loaded")
		}
	} else {
		// Match case of DBF extension
		cdxExt := ".cdx"
		if len(dbfExt) > 0 && dbfExt[0] == '.' {
			if dbfExt[1:2] == "D" || dbfExt[1:2] == "d" {
				if dbfExt[1:2] == "D" {
					cdxExt = ".CDX"
				}
			}
		}
		cdxFilename = baseFilename + cdxExt
	}

	// Open CDX file
	cdxFile, err := dbf.opener.OpenFile(cdxFilename, os.O_RDONLY, 0600)
	if err != nil {
		return NewErrorf("failed to open CDX file %s", cdxFilename).SetWrapped(err).SetContext("ensure cdx loaded")
	}

	// Parse CDX file
	cdx, err := core.OpenCdx(cdxFile)
	if err != nil {
		_ = cdxFile.Close()
		return NewError("failed to parse CDX file").SetWrapped(err).SetContext("ensure cdx loaded")
	}

	dbf.cdx = cdx
	dbf.cdxFilename = cdxFilename

	return nil
}

// Seek searches for a key in the specified CDX tag and positions at the found record
// Returns the record number and seek result
// The searchString is the key value to search for
func (dbf *Dbf) Seek(indexTag string, searchString string) (int32, core.SeekResult, error) {
	if !dbf.Active() {
		return 0, core.SeekError, NewInactiveError().SetContext("seek")
	}

	// Ensure CDX file is loaded
	if err := dbf.ensureCdxLoaded(); err != nil {
		return 0, core.SeekError, NewError("failed to load CDX file").SetWrapped(err).SetContext("seek")
	}

	// Find the specified tag
	tag := dbf.cdx.FindTag(indexTag)
	if tag == nil {
		return 0, core.SeekError, NewError("CDX tag not found: " + indexTag).SetContext("seek")
	}

	// Search for the key
	recNo, result := tag.Seek(dbf.cdx, searchString)

	// If found or positioned after, read the record
	if result == core.SeekSuccess || result == core.SeekAfter {
		// Calculate byte offset and read the record
		byteOffset := int64(dbf.recordoffset + dbf.recordsize*int(recNo-1))
		if _, err := dbf.fl.Seek(byteOffset, 0); err != nil {
			return recNo, result, NewError("failed to seek to record").SetWrapped(err).SetContext("seek")
		}
		if err := dbf.readFunc(); err != nil {
			return recNo, result, NewError("failed to read record").SetWrapped(err).SetContext("seek")
		}
	}

	return recNo, result, nil
}

// Find searches for an exact match in the specified CDX tag
// Returns the record number if found, otherwise 0
// The searchString is the key value to search for
func (dbf *Dbf) Find(indexTag string, searchString string) (int32, error) {
	recNo, result, err := dbf.Seek(indexTag, searchString)
	if err != nil {
		return 0, err
	}

	if result != core.SeekSuccess {
		return 0, NewError("key not found").SetContext("find")
	}

	return recNo, nil
}
