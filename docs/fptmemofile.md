# Memo File Structure (.FPT)

> **Source**: [Microsoft Learn - Visual FoxPro Documentation Archive](https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/8599s21w(v=vs.71))  
> **Date**: October 23, 2006

## Overview

Memo files (**.FPT**) in Visual FoxPro contain one **header record** and any number of **block structures**. The header record contains a pointer to the next free block and the size of the block in bytes. The size is determined by the [SET BLOCKSIZE](d6e1ah7y(v=vs.71)) command when the file is created.

### 🔧 Key Characteristics

- **Header record** starts at file position zero and occupies **512 bytes**
- **Block size** is configurable via `SET BLOCKSIZE` command
- `SET BLOCKSIZE TO 0` sets the block size width to **1 byte**
- **Block positioning** is determined by multiplying the block number by the block size
- All memo blocks start at **even block boundary addresses**
- A memo block can occupy **multiple consecutive blocks**

### 📊 File Structure

Following the header record are the blocks that contain a **block header** and the **text of the memo**. The table file contains block numbers that are used to reference the memo blocks. The position of the block in the memo file is determined by multiplying the block number by the block size (found in the memo file header record).

---

## Memo Header Record

| Byte Offset | Description |
|-------------|-------------|
| **00 – 03** | **Location of next free block**¹ |
| **04 – 05** | Unused |
| **06 – 07** | **Block size** (bytes per block)¹ |
| **08 – 511** | Unused |

¹ *Integers stored with the most significant byte first.*

---

## Memo Block Header and Memo Text

| Byte Offset | Description |
|-------------|-------------|
| **00 – 03** | **Block signature**¹ (indicates the type of data in the block)<br/>• `0` – picture (picture field type)<br/>• `1` – text (memo field type) |
| **04 – 07** | **Length**¹ of memo (in bytes) |
| **08 – n** | **Memo text** (n = length) |

¹ *Integers stored with the most significant byte first.*

---

## Technical Details

### 🗂️ Block Management

- **Free Block Tracking**: The header maintains a pointer to the next available free block
- **Block Allocation**: New memo data is allocated to free blocks or appends to the file
- **Block Reuse**: Deleted memo blocks are added to the free block chain for reuse

### 📝 Data Types Supported

- **Text Data**: Standard memo field content stored as text
- **Picture Data**: Binary picture data for OLE objects and images
- **Variable Length**: Memo blocks can contain data of varying lengths

### 💾 Storage Efficiency

- **Block-based Storage**: Efficient allocation and management of variable-length data
- **Configurable Block Size**: Allows optimization based on typical memo field sizes
- **Boundary Alignment**: Even block boundaries ensure optimal disk access patterns

---

## Usage Notes

### ⚙️ Configuration

The block size is set when the memo file is created and affects:
- **Storage efficiency** for different memo sizes
- **Disk access patterns** and performance
- **File size growth** characteristics

### 🔗 Table Integration

- Memo files work in conjunction with table files (.dbf)
- Table records contain **block numbers** that reference memo blocks
- **Automatic management** of memo block allocation and deallocation

---

## See Also

- [Table File Structure (.dbc, .dbf, .frx, .lbx, .mnx, .pjx, .scx, .vcx)](st4a0s68(v=vs.71))
- [Table Structures of Table Files (.dbc, .frx, .lbx, .mnx, .pjx, .scx, .vcx)](72es52cd(v=vs.71))
- [SET BLOCKSIZE](d6e1ah7y(v=vs.71))
- [Index File Structure (.idx)](x0btabez(v=vs.71))
- [Compact Index File Structure (.idx)](s8tb8f47(v=vs.71))
- [Compound Index File Structure (.cdx)](k35b9hs2(v=vs.71))
- [Macro File Format (.fky)](t711dh3d(v=vs.71))
- [File Extensions and File Types](h9yfa0t1(v=vs.71))

---

*This documentation is part of the archived Microsoft Visual FoxPro documentation and covers the internal structure of memo files (.FPT) used for storing variable-length text and binary data.*