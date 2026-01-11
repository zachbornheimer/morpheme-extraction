// Package filesystem provides abstraction layer for file operations
// Follows Facade Rule from CLAUDE.md for better testability
package filesystem

import (
	"io"
	"os"
)

// FileSystem abstracts file operations for testing and mocking
type FileSystem interface {
	Open(filename string) (ReadCloser, error)
	OpenFile(filename string, flag int, perm os.FileMode) (WriteCloser, error)
	Stdin() Reader
}

// ReadCloser combines io.Reader and io.Closer
type ReadCloser interface {
	io.Reader
	io.Closer
}

// WriteCloser combines io.Writer and io.Closer
type WriteCloser interface {
	io.Writer
	io.Closer
}

// Reader provides Read operation
type Reader interface {
	Read(p []byte) (n int, err error)
}

// OsFileSystem implements FileSystem using standard os package
type OsFileSystem struct{}

// NewOsFileSystem creates a new OsFileSystem
func NewOsFileSystem() *OsFileSystem {
	return &OsFileSystem{}
}

// Open opens a file for reading
func (fs *OsFileSystem) Open(filename string) (ReadCloser, error) {
	return os.Open(filename)
}

// OpenFile opens a file with flags and permissions
func (fs *OsFileSystem) OpenFile(filename string, flag int, perm os.FileMode) (WriteCloser, error) {
	return os.OpenFile(filename, flag, perm)
}

// Stdin returns os.Stdin
func (fs *OsFileSystem) Stdin() Reader {
	return os.Stdin
}
