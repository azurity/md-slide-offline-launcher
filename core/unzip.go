package core

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const exeSize = 12 * 1024 * 1024

type ExeZip struct {
	file *os.File
}

func (file *ExeZip) ReadAt(p []byte, off int64) (n int, err error) {
	return file.file.ReadAt(p, off+exeSize)
}

func (file *ExeZip) Size() (int64, error) {
	fi, err := file.file.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size() - exeSize, nil
}

func Unzip(folder string) error {
	fmt.Println(folder)
	fmt.Println(os.Args[0])
	file, err := os.Open(os.Args[0])
	if err != nil {
		return err
	}
	defer file.Close()
	exeZipFile := &ExeZip{file: file}
	size, err := exeZipFile.Size()
	if err != nil {
		return err
	}
	zipFile, err := zip.NewReader(exeZipFile, size)
	if err != nil {
		return err
	}
	for _, f := range zipFile.File {
		fpath := filepath.Join(folder, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
