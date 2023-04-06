package main

import (
	"context"
	"time"

	"github.com/momentohq/client-sdk-go/momento"
	"github.com/momentohq/client-sdk-go/responses"
	"github.com/prozz/aws-embedded-metrics-golang/emf"
	"go.mongodb.org/mongo-driver/mongo"
)

type handler struct {
	momentoClient momento.CacheClient
	mongoClient   *mongo.Client
	cacheName     string
}

func (h *handler) handle(ctx context.Context) error {
	for queryName, query := range defaultConfig.queriesToRun {

		// First try running queries directly against mongo
		for i := 0; i < defaultConfig.iterationsToRun; i++ {
			startTime := time.Now()
			_, err := query(ctx, h.mongoClient)
			if err != nil {
				return err
			}
			emf.New().MetricAs(
				"mongo-time-taken",
				int(time.Now().Sub(startTime).Microseconds()),
				emf.Microseconds,
			).Dimension("query", string(queryName)).Log()
		}

		// Next load query results directly from momento
		for i := 0; i < defaultConfig.iterationsToRun; i++ {
			startTime := time.Now()

			// Look up item in cache first
			resp, err := h.momentoClient.Get(ctx, &momento.GetRequest{
				CacheName: h.cacheName,
				Key:       momento.String(queryName),
			})
			if err != nil {
				return err
			}

			// Check if we got cache hit or miss
			switch resp.(type) {
			case *responses.GetHit:
				emf.New().MetricAs(
					"momento-time-taken",
					int(time.Now().Sub(startTime).Microseconds()),
					emf.Microseconds,
				).Dimension("query", string(queryName)).Log()
			case *responses.GetMiss:
				//log.Printf("Look up did not find a value key=%s", "cache-key")
				result, err := query(ctx, h.mongoClient)
				if err != nil {
					return err
				}

				// Set result from benchmark into momento so next time there will be a hit
				_, err = h.momentoClient.Set(ctx, &momento.SetRequest{
					CacheName: h.cacheName,
					Key:       momento.String(queryName),
					Value:     momento.String(result),
					Ttl:       time.Second * 60,
				})
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}
