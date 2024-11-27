package sqldb

import (
	"bytes"
	"io"
	"io/fs"
	"time"
)

type oneFileFS struct {
	data []byte
}

func NewOneFileFS(data []byte) fs.FS {
	return &oneFileFS{data}
}

func (f *oneFileFS) Open(name string) (fs.File, error) {
	return &fileObject{name, f.data, bytes.NewReader(f.data)}, nil
}

type fileObject struct {
	name   string
	data   []byte
	reader io.Reader
}

func (f *fileObject) Close() error {
	return nil
}

func (f *fileObject) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *fileObject) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *fileObject) Name() string {
	return f.name
}

func (f *fileObject) Size() int64 {
	return int64(len(f.data))
}

func (f *fileObject) Mode() fs.FileMode {
	return 0
}

func (f *fileObject) ModTime() time.Time {
	return time.Time{}
}

func (f *fileObject) IsDir() bool {
	return false
}

func (f *fileObject) Sys() any {
	return nil
}
