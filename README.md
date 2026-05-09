# foxy

`foxy` is a high-performance Go library for reading and navigating XBase (dBase, FoxPro) DBF files and their associated CDX indexes.

## Features

- **DBF Support:** Read dBase III, IV, and FoxPro DBF files.
- **CDX Support:** Navigate records using FoxPro-style CDX indexes.
- **Fuzzy File Search:** Automatically find associated files (`.dbf`, `.cdx`, `.fpt`) regardless of case sensitivity or naming variations.
- **Custom Openers:** Pluggable file system abstraction allows reading from ZIP files, memory, or custom storage backends.
- **Rich Data Types:** Support for Character, Numeric, Floating, Logical, Date, Memo, and Integer fields.

## Installation

```bash
go get github.com/mkfoss/foxy
```

## Quick Start

### Basic Reading

```go
package main

import (
	"fmt"
	"log"

	"github.com/mkfoss/foxy"
)

func main() {
	dbf := &foxy.Dbf{}
	err := dbf.Open("data/customers.dbf")
	if err != nil {
		log.Fatal(err)
	}
	defer dbf.Close()

	fmt.Printf("Records: %d\n", dbf.RecordCount())

	// Iterate through records in physical order
	for i := 1; i <= dbf.RecordCount(); i++ {
		err := dbf.Goto(i)
		if err != nil {
			log.Fatal(err)
		}

		// Access fields by name
		name := dbf.FieldByName("NAME").MustAsString(true, true, true)
		balance := dbf.FieldByName("BALANCE").MustAsFloat()

		fmt.Printf("Record %d: %s - $%.2f\n", i, name, balance)
	}
}
```

### Indexed Navigation

```go
dbf := &foxy.Dbf{}
// Opening with a CDX index
err := dbf.Open("data/customers.dbf")
if err != nil {
    log.Fatal(err)
}

// Set a navigator to use a specific index tag
nav := foxy.NewCdxNavigator("CUSTNAME")
dbf.SetNavigator(nav)

// Now First(), Next(), etc. follow the index order
for err := dbf.First(); err == nil; err = dbf.Next() {
    fmt.Println(dbf.FieldByName("NAME").MustAsString(true, true, true))
}
```

## Advanced Usage

### Fuzzy File Search

`foxy` can automatically locate associated files even if they have different casing or are in different directories (useful when working with legacy data on Linux).

```go
dbf := &foxy.Dbf{UseFuzzyFileSearch: true}
err := dbf.Open("data/CUSTOMERS.DBF") 
// Will find customers.dbf, customers.cdx, and customers.fpt automatically
```

### Custom Openers

Implement the `Opener` interface to read files from non-standard sources like ZIP archives or S3 buckets.

## License

MIT License. See [LICENSE](LICENSE) for details.
