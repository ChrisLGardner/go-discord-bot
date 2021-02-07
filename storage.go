package main

import (
	"context"

	"github.com/honeycombio/beeline-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect to the specified mongo instance using the context for timeout
func connectDb(ctx context.Context, uri string) (*mongo.Client, error) {
	ctx, span := beeline.StartSpan(ctx, "mongo.connect")
	defer span.Send()

	span.AddField("mongo.server", uri)

	clientOptions := options.Client().ApplyURI(uri).SetDirect(true)
	c, err := mongo.NewClient(clientOptions)
	if err != nil {
		span.AddField("mongo.client.error", err)
		return nil, err
	}

	err = c.Connect(ctx)
	if err != nil {
		span.AddField("mongo.connect.error", err)
		return nil, err
	}

	err = c.Ping(ctx, nil)
	if err != nil {
		span.AddField("mongo.ping.error", err)
		return nil, err
	}

	return c, nil
}

func runQuery(ctx context.Context, mc *mongo.Client, query interface{}) ([]bson.M, error) {

	ctx, span := beeline.StartSpan(ctx, "mongo.runQuery")
	defer span.Send()

	collection := mc.Database("reminders").Collection("reminders")
	span.AddField("mongo.runQuery.collection", collection.Name())
	span.AddField("mongo.runQuery.database", collection.Database().Name())
	span.AddField("mongo.runQuery.query", query)

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		span.AddField("mongo.runQuery.error", err)
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		span.AddField("mongo.runQuery.error", err)
		return nil, err
	}

	span.AddField("mongo.runQuery.results.Count", len(results))
	span.AddField("mongo.runQuery.results.raw", results)

	return results, nil
}

func writeDbObject(ctx context.Context, mc *mongo.Client, obj interface{}) error {

	ctx, span := beeline.StartSpan(ctx, "mongo.writeObject")
	defer span.Send()

	data, err := bson.Marshal(obj)
	if err != nil {
		span.AddField("mongo.writeObject.error", err)
		return err
	}

	collection := mc.Database("reminders").Collection("reminders")
	span.AddField("mongo.writeObject.collection", collection.Name())
	span.AddField("mongo.writeObject.database", collection.Database().Name())
	span.AddField("mongo.writeObject.object", data)

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		span.AddField("mongo.writeObject.error", err)
		return err
	}

	span.AddField("mongo.writeObject.id", res.InsertedID)

	return nil
}
