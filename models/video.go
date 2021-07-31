package models

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
)

type (
	VideoFileState int

	//Video represents a video file stored on the server
	Video struct {
		Title       string         `bson:"title" json:"title"`
		FileName    string         `bson:"file_name" json:"file_name"`
		Description string         `bson:"description" json:"description"`
		VideoURL    string         `bson:"url" json:"url"`
		AvatarImg   string         `bson:"avatar_img" json:"avatar_img"`
		Hash        string         `bson:"hash" json:"hash"`
		Chunks      []VideoChunk   `bson:"chunks" json:"chunks"`
		Size        int64          `bson:"size" json:"size"`
		State       VideoFileState `bson:"state"`
		mtx         sync.Mutex     `bson:"-" json:"-"`
	}

	VideoMetaResponse struct {
		Title       string `bson:"title" json:"title"`
		Description string `bson:"description" json:"description"`
		VideoURL    string `bson:"url" json:"url"`
		avatarImg   string `bson:"avatar_img" json:"avatar_img"`
	}
)

const (
	Incomplete VideoFileState = iota // Not all chunks are received
	Complete                         // All chunks are received
	Merged                           // All chunks are merged to single file and checksum matches
)

func NewVideo(title string, fileName string, description string, hash string, chunks []VideoChunk, size int64) *Video {
	return &Video{Title: title, FileName: fileName, Description: description, VideoURL: fmt.Sprintf("storage/%x/%s", hash, fileName),
		Hash: hash, Chunks: chunks, Size: size, State: Incomplete}
}

func (v *Video) GetStorageDirectory() string {
	return fmt.Sprintf("storage/%s", v.Hash)
}

func (v *Video) GetStorageFilePath() string {
	return fmt.Sprintf("storage/%s/%s", v.Hash, v.FileName)
}

func (v *Video) getState() VideoFileState {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	return v.State
}

func (v *Video) setState(state VideoFileState) {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	v.State = state
}

//IsAllChunkReceived returns whether all video's chunks has been saved on disk
func (v *Video) IsAllChunkReceived() bool {
	if v.getState() != Incomplete {
		return true
	}
	for _, chunk := range v.Chunks {
		if !Storage.IsFileExists(chunk.GetChunkFilePath()) {
			return false
		}
	}
	// If all chunk has been received, change the state to Complete
	v.setState(Complete)
	return true
}

//DeleteAllChunks deletes video's all chunks saved on the disk
func (v *Video) DeleteAllChunks() error {
	for _, chunk := range v.Chunks {
		err := Storage.DeleteFileIfExists(chunk.GetChunkFilePath())
		if err != nil {
			return err
		}
	}
	return nil
}

//MergeChunks merges video's all chunks if video's state is Complete, also delete all chunks data after merging
func (v *Video) MergeChunks() error {
	if !v.IsAllChunkReceived() {
		return errors.New("chunks incomplete!")
	}

	file, err := os.OpenFile(v.GetStorageFilePath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	for _, chunk := range v.Chunks {
		content, err := chunk.ReadDiskContent()
		if err != nil {
			return err
		}
		n, err := file.Write(content)
		if err != nil || n != len(content) {
			return errors.New("fail to Write chunk content to merged file")
		}
		err = file.Sync()
		if err != nil {
			return err
		}
		content = nil
	}
	file.Close()
	// After all chunks has been merged
	// 1. check the merged file's md5 must match our record
	md5, err := Storage.ComputeFileMD5(v.GetStorageFilePath())
	if err != nil {
		return err
	}
	if hex.EncodeToString(md5) != v.Hash {
		// If the md5 hash does not match, we need to delete all chunks and the merged file, and redo upload
		// (I hope this scenario never happens)
		if err := Storage.DeleteFileIfExists(v.GetStorageFilePath()); err != nil {
			return err
		}
		if err := v.DeleteAllChunks(); err != nil {
			return err
		}
		return errors.New("the merged file's hash does not match our record")
	}
	// 2. set the file's state to be merged
	v.setState(Merged)
	// 3. delete all chunk files
	if err := v.DeleteAllChunks(); err != nil {
		return err
	}
	return nil
}
