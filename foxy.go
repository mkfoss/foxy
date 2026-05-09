/*
Package foxy is a high-performance library for reading and navigating XBase (dBase, FoxPro) DBF files
and their associated CDX indexes.

It provides a rich set of features including support for multiple data types (Character, Numeric,
Logical, Date, Memo, etc.), fuzzy file searching to handle case-sensitivity issues on modern
file systems, and a pluggable file system abstraction (Opener) for reading from non-standard sources.

Basic Usage:

	dbf := &foxy.Dbf{}
	err := dbf.Open("data/customers.dbf")
	if err != nil {
		log.Fatal(err)
	}
	defer dbf.Close()

	for i := 1; i <= dbf.RecordCount(); i++ {
		dbf.Goto(i)
		fmt.Println(dbf.FieldByName("NAME").MustAsString(true, true, true))
	}

Indexed Navigation:

	nav := foxy.NewCdxNavigator("CUSTNAME")
	dbf.SetNavigator(nav)

	for err := dbf.First(); err == nil; err = dbf.Next() {
		fmt.Println(dbf.FieldByName("NAME").MustAsString(true, true, true))
	}
*/
package foxy
