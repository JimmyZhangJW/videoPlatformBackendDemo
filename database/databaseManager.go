package database

import (
	"context"
	"github.com/JimmyZhangJW/videoPlatformBackendDemo/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Manager struct {
	client *mongo.Client
}

var Database = &Manager{}

// establish connection to the database on init
func init() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:pass12345@localhost:27017"))
	if err != nil {
		log.Fatalln(err)
	}
	Database.client = client
}

func (db *Manager) InsertVideoMeta(ctx context.Context, vm models.Video) error {
	collection := db.client.Database("video-platform").Collection("videoMetas")
	_, err := collection.InsertOne(ctx, vm)
	return err
}

func (db *Manager) GetAllPublicVideoMetas(ctx context.Context) ([]models.VideoMetaResponse, error) {
	var response []models.VideoMetaResponse
	collection := db.client.Database("video-platform").Collection("videoMetas")
	cursor, err := collection.Find(ctx, bson.D{{"state", models.Merged}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func (db *Manager) GetVideoMetaWithHash(ctx context.Context, hash string) (*models.Video, error) {
	var response models.Video
	collection := db.client.Database("video-platform").Collection("videoMetas")
	one := collection.FindOne(ctx, bson.D{{"hash", hash}})
	err := one.Decode(&response)
	return &response, err
}

func (db *Manager) CheckVideoMetaExists(ctx context.Context, hash string) (bool, error) {
	collection := db.client.Database("video-platform").Collection("videoMetas")
	count, err := collection.CountDocuments(ctx, bson.D{{"hash", hash}})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *Manager) UpdateVideoMetaState(ctx context.Context, hash string, state int) error {
	opts := options.Update().SetUpsert(true)
	collection := db.client.Database("video-platform").Collection("videoMetas")
	_, err := collection.UpdateOne(ctx, bson.D{{"hash", hash}}, bson.D{{"$set", bson.D{{"state", state}}}}, opts)
	return err
}
