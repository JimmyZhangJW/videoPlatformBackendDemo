package models

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

//VideoChunk represents a slice of video
type VideoChunk struct {
	Index    int    `bson:"index" json:"index"`
	Hash     []byte `bson:"hash" json:"hash"`
	Size     int64  `bson:"size" json:"size"`
	FileName string `bson:"file_name" json:"file_name"`
	FileHash []byte `bson:"file_hash" json:"file_hash"`
}

var (
	ParentFileFolderNotExist = errors.New("parent file's folder does not exist")
	FileChunkNotMatch        = errors.New("chunk's hash or size does not match our record")
	ChunkNotOnDisk           = errors.New("chunk is not on the disk")
)

//Constructor: NewVideoChunk
func NewVideoChunk(index int, hash []byte, size int64, fileName string, fileHash []byte) *VideoChunk {
	return &VideoChunk{Index: index, Hash: hash, Size: size, FileName: fileName, FileHash: fileHash}
}

func (vc *VideoChunk) GetChunkFilePath() string {
	return fmt.Sprintf("storage/%x/%x", vc.FileHash, vc.Hash)
}

//StoreToDisk saves chunk data to the disk only if the content's hash and size matches the VideoChunk's record
func (vc *VideoChunk) StoreToDisk(content []byte) error {
	//1. verify the content's hash and size
	sum := md5.Sum(content)
	// If hash or size does not match, return FileChunkHashNotMatch error
	if !bytes.Equal(sum[:], vc.Hash) || int64(len(content)) != vc.Size {
		return FileChunkNotMatch
	}
	//2. store the content to the disk
	return storage.SaveFile("", content)
}

//ReadDiskContent reads the chunk's bytes from the disk
func (vc *VideoChunk) ReadDiskContent() ([]byte, error) {
	if !storage.IsFileExists(vc.GetChunkFilePath()) {
		return nil, ChunkNotOnDisk
	}
	file, err := os.Open(vc.GetChunkFilePath())
	if err != nil {
		return nil, err
	}
	all, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return all, nil
}
