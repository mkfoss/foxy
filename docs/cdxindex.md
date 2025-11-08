# Compact Index File Structure (.idx)

> **Source**: [Microsoft Learn - Visual FoxPro Documentation Archive](https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/s8tb8f47(v=vs.71))  
> **Date**: October 23, 2006

## Overview

This document describes the internal structure of Visual FoxPro compact index files (**.idx**). Compact index files are used to store index information in a more space-efficient format compared to standard index files.

### Compound Index File Structure (.cdx)

All compound indexes are compact indexes.

One file structure exists to track all the tags in the .cdx file. This structure is identical to the compact index file structure with one exception — the leaf nodes at the lowest level of this structure point to one of the tags in the compound index.

All tags in the index have their own complete structure that is identical to the compact index structure for an .idx file.
---

## Compact Index Header Record

| Byte Offset | Description |
|-------------|-------------|
| **00 – 03** | Pointer to root node |
| **04 – 07** | Pointer to free node list (`-1` if not present) |
| **08 – 11** | Reserved for internal use |
| **12 – 13** | Length of key |
| **14** | **Index options** (any of the following numeric values or their sums):<br/>• `1` – a unique index<br/>• `8` – index has FOR clause<br/>• `32` – compact index format<br/>• `64` – compound index header |
| **15** | Index signature |
| **16 – 19** | Reserved for internal use |
| **20 – 23** | Reserved for internal use |
| **24 – 27** | Reserved for internal use |
| **28 – 31** | Reserved for internal use |
| **32 – 35** | Reserved for internal use |
| **36 – 501** | Reserved for internal use |
| **502 – 503** | **Ascending or descending:**<br/>• `0` = ascending<br/>• `1` = descending |
| **504 – 505** | Reserved for internal use |
| **506 – 507** | FOR expression pool length¹ |
| **508 – 509** | Reserved for internal use |
| **510 – 511** | Key expression pool length¹ |
| **512 – 1023** | Key expression pool (uncompiled) |

¹ *This information tracks the space used in the key expression pool.*

---

## Compact Index Interior Node Record

| Byte Offset | Description |
|-------------|-------------|
| **00 – 01** | **Node attributes** (any of the following numeric values or their sums):<br/>• `0` – index node<br/>• `1` – root node<br/>• `2` – leaf node |
| **02 – 03** | Number of keys present (0, 1 or many) |
| **04 – 07** | Pointer to node directly to left of current node (on same level, `-1` if not present) |
| **08 – 11** | Pointer to node directly to right of current node (on same level; `-1` if not present) |
| **12 – 511** | **Up to 500 characters** containing the key value for the length of the key with a four-byte hexadecimal number (stored in normal left-to-right format)<br/><br/>ℹ️ This node always contains the index key, record number and intra-index pointer.²<br/><br/>The key/four-byte hexadecimal number combinations will occur the number of times indicated in bytes 02 – 03. |

---

## Compact Index Exterior Node Record

| Byte Offset | Description |
|-------------|-------------|
| **00 – 01** | **Node attributes** (any of the following numeric values or their sums):<br/>• `0` – index node<br/>• `1` – root node<br/>• `2` – leaf node |
| **02 – 03** | Number of keys present (0, 1 or many) |
| **04 – 07** | Pointer to the node directly to the left of current node (on same level; `-1` if not present) |
| **08 – 11** | Pointer to the node directly to right of the current node (on same level; `-1` if not present) |
| **12 – 13** | Available free space in node |
| **14 – 17** | Record number mask |
| **18** | Duplicate byte count mask |
| **19** | Trailing byte count mask |
| **20** | Number of bits used for record number |
| **21** | Number of bits used for duplicate count |
| **22** | Number of bits used for trail count |
| **23** | Number of bytes holding record number, duplicate count and trailing count |
| **24 – 511** | Index keys and information² |

---

## Technical Notes

### 📝 Entry Structure

² Each entry consists of the record number, duplicate byte count and trailing byte count, all compacted. The key text is placed at the logical end of the node, working backwards, allowing for previous key entries.

### 🔍 Key Features

- **Compact Format**: Designed to minimize storage space while maintaining index functionality
- **Node Types**: Support for different node types (index, root, leaf) for efficient tree traversal  
- **Bidirectional Links**: Nodes maintain pointers to left and right neighbors for efficient navigation
- **Expression Pools**: Dedicated space for storing key and FOR clause expressions

---

## See Also

- [Compound Index File Structure (.cdx)](k35b9hs2(v=vs.71))
- [Index File Structure (.idx)](x0btabez(v=vs.71))
- [Table File Structure (.dbc, .dbf, .frx, .lbx, .mnx, .pjx, .scx, .vcx)](st4a0s68(v=vs.71))
- [Table Structures of Table Files (.dbc, .frx, .lbx, .mnx, .pjx, .scx, .vcx)](72es52cd(v=vs.71))
- [Memo File Structure (.FPT)](8599s21w(v=vs.71))
- [Macro File Format (.fky)](t711dh3d(v=vs.71))
- [File Extensions and File Types](h9yfa0t1(v=vs.71))

---

*This documentation is part of the archived Microsoft Visual FoxPro documentation and covers the internal structure of compact index files (.idx) used for efficient data indexing.*