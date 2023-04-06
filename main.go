package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/momentohq/client-sdk-go/auth"
	momentoConfig "github.com/momentohq/client-sdk-go/config"
	"github.com/momentohq/client-sdk-go/momento"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		cacheName = os.Getenv("CACHE_NAME")
		mongoURI  = os.Getenv("MONGO_URI")
	)

	if cacheName == "" {
		log.Fatal(errors.New("CACHE_NAME undefined"))
	}
	if mongoURI == "" {
		log.Fatal(errors.New("MONGO_URI undefined"))
	}

	// Load Momento Auth Token From Environment Variable
	var credentialProvider, err = auth.FromEnvironmentVariable("MOMENTO_AUTH_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// Initializes Momento
	momClient, err := momento.NewCacheClient(
		momentoConfig.InRegionLatest(),
		credentialProvider,
		60*time.Second,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Make dummy get request against cache to establish eager connection to momento
	// TODO pull this down to SDK momentoConfig setting on init to force to happen
	_, err = momClient.Get(context.Background(), &momento.GetRequest{
		CacheName: cacheName,
		Key:       momento.String("warmup"),
	})

	// Init Mongo DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// establish eager connection to mongo
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create handler and kick off benchmark queries defined in `config.go
	h := handler{
		momentoClient: momClient,
		mongoClient:   client,
		cacheName:     cacheName,
	}
	err = h.handle(context.Background())
	if err != nil {
		log.Fatal(err)
	}

}
