package clients

import (
	"io"
	"os"
)

type IFilesystem interface {
	Open(name string) (IFile, error)
	Create(name string) (IFile, error)
	Remove(name string) error
}

type IFile interface {
	io.Writer
	io.WriterAt
	io.Reader
	io.Closer
}

type OsFS struct{}

func (OsFS) Open(name string) (IFile, error)   { return os.Open(name) }
func (OsFS) Create(name string) (IFile, error) { return os.Create(name) }
func (OsFS) Remove(name string) error          { return os.Remove(name) }
