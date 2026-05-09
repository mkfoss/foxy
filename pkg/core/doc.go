/*
Package core provides low-level primitives for working with XBase (dBase, FoxPro) file formats.

This package is designed for developers who need to interact with DBF, CDX, and FPT files at a
granular level without using the higher-level abstractions provided by the root foxy package.

It includes:
  - Header and Field definition parsing
  - Record data manipulation
  - CDX index B-tree traversal and iterators
  - File system abstractions (Opener) for flexible storage backends
  - Fuzzy file search utilities for cross-platform filename compatibility
*/
package core
