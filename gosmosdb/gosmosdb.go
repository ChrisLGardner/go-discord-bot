package gosmosdb

import (
	"context"
)

type DbReader interface {
	Find(ctx context.Context, filter interface{}) (result []interface{}, err error)
}

type DbWriter interface {
	InsertOne(ctx context.Context, document interface{}) (result interface{}, err error)
}
