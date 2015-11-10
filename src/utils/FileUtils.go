package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func FileReadBytes(path string) ([]byte, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if finfo.Size() == 0 {
		return nil, fmt.Errorf("size is zero. file: %s", path)
	}

	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	return ioutil.ReadAll(fi)
}

func Mkdirp(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
			return false, err
		} else {
			return true, nil
		}
	} else {
		return false, err
	}
}

func FileCopy(destPath string, srcPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, srcFile)
	return err
}
