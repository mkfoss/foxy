# Table File Structure (.dbc, .dbf, .frx, .lbx, .mnx, .pjx, .scx, .vcx)

> **Source**: [Microsoft Learn - Visual FoxPro Documentation Archive](https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/st4a0s68(v=vs.71))  
> **Date**: October 23, 2006

## Overview

Visual FoxPro uses tables to store data that defines different file types. The file types that are saved as table files are:

- **Table** (.dbf)
- **Database** (.dbc)
- **Form** (.scx)
- **Label** (.lbx)
- **Menu** (.mnx)
- **Project** (.pjx)
- **Report** (.frx)
- **Visual Class Library** (.vcx)

Because these files are actually tables, you can use and browse them in the same way that you browse any .dbf file.

A table file is made up of a **header record** and **data records**. The header record defines the structure of the table and contains any other information related to the table. It starts at file position zero. The data records¹ follow the header (in consecutive bytes) and contain the actual text of the fields.

> 📚 **Related**: For information about the table structures of the different file types, see [Table Structures of Table Files](72es52cd(v=vs.71)).

The length of a record (in bytes) is determined by summing the defined lengths of all fields. Integers in table files are stored with the least significant byte first.

---

## <a name="dbfheader">Table Header Record Structure</a>

| Byte Offset | Description |
|-------------|-------------|
| **0** | **Type of file**<br/>• `0x02` - FoxBASE<br/>• `0x03` - FoxBASE+/dBASE III PLUS, no memo<br/>• `0x30` - Visual FoxPro<br/>• `0x43` - dBASE IV SQL table files, no memo<br/>• `0x63` - dBASE IV SQL system files, no memo<br/>• `0x83` - FoxBASE+/dBASE III PLUS, with memo<br/>• `0x8B` - dBASE IV with memo<br/>• `0xCB` - dBASE IV SQL table files, with memo<br/>• `0xF5` - FoxPro 2.x (or earlier) with memo<br/>• `0xFB` - FoxBASE |
| **1 – 3** | Last update (YYMMDD) |
| **4 – 7** | Number of records in file |
| **8 – 9** | Position of first data record |
| **10 – 11** | Length of one data record (including delete flag) |
| **12 – 27** | Reserved |
| **28** | **Table Flags**<br/>• `0x01` - file has a structural .cdx<br/>• `0x02` - file has a Memo field<br/>• `0x04` - file is a database (.dbc)<br/><br/>⚠️ **Note**: This byte can contain the sum of any of the above values. For example, `0x03` indicates the table has a structural .cdx and a Memo field. |
| **29** | Code page mark |
| **30 – 31** | Reserved, contains `0x00` |
| **32 – n** | **Field subrecords**<br/>The number of fields determines the number of field subrecords. There is one field subrecord for each field in the table. |
| **n+1** | Header record terminator (`0x0D`) |
| **n+2 to n+264** | A 263-byte range that contains the backlink information (the relative path of an associated database (.dbc)). If the first byte is `0x00` then the file is not associated with a database. Hence, database files themselves always contain `0x00`. |

---

## Field Subrecords Structure

| Byte Offset | Description |
|-------------|-------------|
| **0 – 10** | Field name (maximum of 10 characters; if less than 10, it is padded with null character (`0x00`)) |
| **11** | **Field Type:**<br/>• `C` – Character<br/>• `Y` – Currency<br/>• `N` – Numeric<br/>• `F` – Float<br/>• `D` – Date<br/>• `T` – DateTime<br/>• `B` – Double<br/>• `I` – Integer<br/>• `L` – Logical<br/>• `M` – Memo<br/>• `G` – General<br/>• `C` – Character (binary)<br/>• `M` – Memo (binary)<br/>• `P` – Picture |
| **12 – 15** | Displacement of field in record |
| **16** | Length of field (in bytes) |
| **17** | Number of decimal places |
| **18** | **Field Flags**<br/>• `0x01` - System Column (not visible to user)<br/>• `0x02` - Column can store null values<br/>• `0x04` - Binary column (for CHAR and MEMO only) |
| **19 – 32** | Reserved |

---

## Data Records

¹ The data in the data file starts at the position indicated in bytes 8 to 9 of the header record. Data records begin with a **delete flag byte**:

- **ASCII space** (`0x20`) = record is **not deleted**
- **Asterisk** (`0x2A`) = record is **deleted**

The data from the fields named in the field subrecords follows the delete flag.

---

## Remarks

Visual FoxPro does not modify the header of a file that has been saved to a FoxPro 2.*x* file format unless one of the following features has been added to the file:

- Null value support
- DateTime, Currency, and Double data types  
- CHAR or MEMO field is marked as Binary
- A table is added to a database (.dbc) file

### 💡 Field Count Formula

**Tip**: You can use the following formula to return the number of fields in a table file:

```
(x – 296) / 32
```

Where:
- `x` = position of the first record (bytes 8 to 9 in the table header record)
- `296` = 263 (backlink info) + 1 (header record terminator) + 32 (first field subrecord)  
- `32` = length of a field subrecord

---

## See Also

- [Code Pages Supported by Visual FoxPro](8t45x02s(v=vs.71))
- [Data and Field Types](ww305zh2(v=vs.71))
- [Table Structures of Table Files (.dbc, .frx, .lbx, .mnx, .pjx, .scx, .vcx)](72es52cd(v=vs.71))
- [Visual FoxPro System Capacities](3kfd3hw9(v=vs.71))
- [Checking for Differences in Forms, Reports, and Other Table Files](eb43dww2(v=vs.71))
- [File Extensions and File Types](h9yfa0t1(v=vs.71))

---

*This documentation is part of the archived Microsoft Visual FoxPro documentation and covers the internal structure of Visual FoxPro table files.*