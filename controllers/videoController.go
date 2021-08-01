package controllers

import (
	"context"
	"fmt"
	"github.com/JimmyZhangJW/videoPlatformBackendDemo/database"
	"github.com/JimmyZhangJW/videoPlatformBackendDemo/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type VideoController struct{}

var VideoC = &VideoController{}

func (vc *VideoController) PostVideoMetaData(c *gin.Context) {
	var videoMeta models.Video
	err := c.ShouldBindJSON(&videoMeta)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Fail to parse JSON data" + err.Error(),
		})
		return
	}
	videoMeta.FileName = strings.ReplaceAll(videoMeta.FileName, " ", "")
	videoMeta.VideoURL = "http://localhost:8080/" + videoMeta.GetStorageFilePath()

	exists, err := database.Database.CheckVideoMetaExists(context.Background(), videoMeta.Hash)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Fail to check if videoMeta exists in the database" + err.Error(),
		})
		return
	}
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"continue": 0,
			"message":  "The file has already exists.",
		})
		return
	}

	if err = models.Storage.CreateFolderIfNotExists(videoMeta.GetStorageDirectory()); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Fail to create folder in storage" + err.Error(),
		})
		return
	}

	if err = database.Database.InsertVideoMeta(context.Background(), videoMeta); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Fail to insert into database" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"continue": 1,
		"message":  "The videoMeta has been uploaded",
	})
}

func (vc *VideoController) PostVideoChunk(c *gin.Context) {
	index, err := strconv.Atoi(c.PostForm("index"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "index is not available",
		})
		return
	}

	videoChunk := models.NewVideoChunk(index, c.PostForm("hash"), c.PostForm("file_hash"))
	// If the corresponding file already exists, return success directly
	if models.Storage.IsFileExists(videoChunk.GetChunkFilePath()) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Your file already exists.",
		})
		return
	}

	file, err := c.FormFile("content")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file is received",
		})
		return
	}

	err = c.SaveUploadedFile(file, fmt.Sprintf("storage/%s/%s", videoChunk.FileHash, videoChunk.Hash))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to save the file: " + err.Error(),
		})
		return
	}

	if err := videoChunk.VerifyDiskContent(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to save the file: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Your file has been successfully uploaded.",
	})
}

func (vc *VideoController) Merge(c *gin.Context) {
	hash := c.PostForm("hash")
	video, err := database.Database.GetVideoMetaWithHash(context.Background(), hash)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to retrieve the video data: " + err.Error(),
		})
		return
	}

	if err = video.MergeChunks(); err != nil {
		// If there is an error, will retry and merge again
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	if err = database.Database.UpdateVideoMetaState(context.Background(), hash, int(models.Merged)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	// If there is no error, it means the merge is successful!
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (vc *VideoController) GetPublicVideoMeta(c *gin.Context) {
	videoMetas, err := database.Database.GetAllPublicVideoMetas(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"Message": "cannot fetch database",
		})
	}
	c.JSON(http.StatusOK, videoMetas)
}
