package gomongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReadOneResult[T any] struct {
	Found    bool
	Document T
	Err      error
	DbRes    *mongo.SingleResult
}

type ReadManyResult[T any] struct {
	Documents []T
	Err       error
}

type ReadStreamResult[T any] struct {
	DocumentStream chan ReadOneResult[T]
	Err            error
}

type DistinctResult[T any] struct {
	Values []T
	Err    error
}

type WriteManyResult struct {
	DbRes mongo.InsertManyResult
	Err   error
}

type WriteOneResult struct {
	DbRes *mongo.InsertOneResult
	Err   error
}

type UpdateResult struct {
	DbRes *mongo.UpdateResult
	Err   error
}

type BulkWriteResult struct {
	Err   error
	DbRes *mongo.BulkWriteResult
}

type DeleteResult struct {
	DelCount int64
	Err      error
}

type CountResult struct {
	Count int64
	Err   error
}

type CommandResult struct {
	DbRes *mongo.SingleResult
	Err   error
}

type IndexCreateResult struct {
	IndexName string
	Err       error
}

type IndexDropResult struct {
	Doc bson.Raw
	Err error
}

type IndexListResult struct {
	Result []interface{}
	Err    error
}
