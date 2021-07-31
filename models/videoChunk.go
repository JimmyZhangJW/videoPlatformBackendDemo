package models

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

//VideoChunk represents a slice of video
type VideoChunk struct {
	Index    int    `bson:"index" json:"index"`
	Hash     string `bson:"hash" json:"hash"`
	FileHash string `bson:"file_hash" json:"file_hash"`
}

var (
	FileChunkNotMatch        = errors.New("chunk's hash or size does not match our record")
	ChunkNotOnDisk           = errors.New("chunk is not on the disk")
)

//Constructor: NewVideoChunk
func NewVideoChunk(index int, hash string, fileHash string) *VideoChunk {
	return &VideoChunk{Index: index, Hash: hash, FileHash: fileHash}
}

func (vc *VideoChunk) GetChunkFilePath() string {
	return fmt.Sprintf("storage/%s/%s", vc.FileHash, vc.Hash)
}

//StoreToDisk saves chunk data to the disk only if the content's hash and size matches the VideoChunk's record
func (vc *VideoChunk) StoreToDisk(content []byte) error {
	//1. verify the content's hash and size
	sum := md5.Sum(content)
	// If hash or size does not match, return FileChunkHashNotMatch error
	if hex.EncodeToString(sum[:]) != vc.Hash {
		return FileChunkNotMatch
	}
	//2. store the content to the disk
	return Storage.SaveFile("", content)
}

//ReadDiskContent reads the chunk's bytes from the disk
func (vc *VideoChunk) ReadDiskContent() ([]byte, error) {
	if !Storage.IsFileExists(vc.GetChunkFilePath()) {
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

func (vc *VideoChunk) VerifyDiskContent() error {
	content, err := vc.ReadDiskContent()
	if err != nil {
		return err
	}
	sum := md5.Sum(content)
	// If hash does not match, return FileChunkHashNotMatch error
	if hex.EncodeToString(sum[:]) != vc.Hash {
		log.Println(hex.EncodeToString(sum[:]), vc.Hash)
		_ = Storage.DeleteFileIfExists(vc.GetChunkFilePath())
		return FileChunkNotMatch
	}
	return nil
}
