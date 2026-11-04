package progress

import (
	"os"
	"sync"
)

type File struct {
	mux        sync.Mutex
	total      int
	downloaded int
	file       *os.File
	onProgress func(float64)
}

func NewFile(total int, f *os.File, onProgress func(float64)) *File {
	return &File{
		total:      total,
		file:       f,
		onProgress: onProgress,
	}
}

func (f *File) Write(b []byte) (int, error) {
	n, err := f.file.Write(b)
	if err == nil {
		f.mux.Lock()
		defer f.mux.Unlock()
		f.downloaded += n
		if f.total > 0 && f.onProgress != nil {
			f.onProgress(float64(f.downloaded) / float64(f.total))
		}
	}
	return n, err
}

func (f *File) WriteAt(b []byte, off int64) (int, error) {
	n, err := f.file.WriteAt(b, off)
	if err == nil {
		f.mux.Lock()
		defer f.mux.Unlock()
		f.downloaded += n
		if f.total > 0 && f.onProgress != nil {
			f.onProgress(float64(f.downloaded) / float64(f.total))
		}
	}
	return n, err
}

func (f *File) Close() error {
	return f.file.Close()
}
