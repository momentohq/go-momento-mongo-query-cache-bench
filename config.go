package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MomentoAuthToken = ""

type queryName string
type mongoQuery func(ctx context.Context, client *mongo.Client) (string, error)
type config struct {
	iterationsToRun int
	queriesToRun    map[queryName]mongoQuery
}

var defaultConfig = config{
	iterationsToRun: 5,
	queriesToRun: map[queryName]mongoQuery{
		"200-items": func(ctx context.Context, client *mongo.Client) (string, error) {
			find, err := client.Database("sample_restaurants").
				Collection("restaurants").
				Find(ctx, bson.D{}, options.Find().SetLimit(200))
			if err != nil {
				return "", err
			}
			var results []bson.M
			err = find.All(ctx, &results)
			if err != nil {
				return "", err
			}
			return fmt.Sprintln(results), nil
		},
		"400-items": func(ctx context.Context, client *mongo.Client) (string, error) {
			find, err := client.Database("sample_restaurants").
				Collection("restaurants").
				Find(ctx, bson.D{}, options.Find().SetLimit(400))
			if err != nil {
				return "", err
			}
			var results []bson.M
			err = find.All(ctx, &results)
			if err != nil {
				return "", err
			}
			return fmt.Sprintln(results), nil
		},
	},
}
