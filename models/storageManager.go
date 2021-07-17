package models

import (
	"crypto/md5"
	"errors"
	"io/ioutil"
	"os"
)

//StorageManager handles troublesome storage-related tasks
type StorageManager struct{}

var storage = &StorageManager{}

var (
	FileAlreadyExisted = errors.New("target file already exists")
	PartialWriteError  = errors.New("only partial file has been written to disk")
)

//IsFolderExists checks whether there is a folder as referenced by the given path
func (storage *StorageManager) IsFolderExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}

//IsFileExists checks whether there is a file as referenced by the given path
func (storage *StorageManager) IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		return true
	}
	return false
}

//CreateFolder creates a folder at the given path
func (storage *StorageManager) CreateFolder(path string) error {
	return os.Mkdir(path, os.ModePerm)
}

//SaveFile saves the content as a file to the given path
func (storage *StorageManager) SaveFile(path string, content []byte) error {
	if storage.IsFileExists(path) {
		return FileAlreadyExisted
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	n, err := file.Write(content)
	if n != len(content) {
		return PartialWriteError
	}
	if err != nil {
		return err
	}
	return file.Sync()
}

//DeleteFileIfExists deletes the file specified by the given path if it exists
func (storage *StorageManager) DeleteFileIfExists(path string) error {
	if storage.IsFileExists(path) {
		return os.Remove(path)
	}
	return nil
}

//ComputeFileMD5 reads content from the given file and return the computed md5 hash
func (storage *StorageManager) ComputeFileMD5(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	sum := md5.Sum(content)
	return sum[:], nil
}
