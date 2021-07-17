package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:pass12345@localhost:27017"))
	if err != nil {
		log.Fatalln(err)
	}

	collection := client.Database("video-platform").Collection("videos")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatalln(err)
	}

	var videos []bson.M
	err = cursor.All(context.Background(), &videos)
	if err != nil {
		log.Fatalln(err)
	}
	for _, video := range videos {
		log.Println(video)
	}

}
