package gomongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GenericResult interface {
}

type ReadOneResult[T any] struct {
	GenericResult
	Found    bool
	Document T
	Err      error
	DbRes    *mongo.SingleResult
}

type ReadManyResult[T any] struct {
	GenericResult
	Documents []T
	Err       error
}

type ReadStreamResult[T any] struct {
	GenericResult
	DocumentStream chan ReadOneResult[T]
	Err            error
}

type WriteManyResult struct {
	GenericResult
	DbRes mongo.InsertManyResult
	Err   error
}

type WriteOneResult struct {
	GenericResult
	DbRes *mongo.InsertOneResult
	Err   error
}

type UpdateResult struct {
	GenericResult
	DbRes *mongo.UpdateResult
	Err   error
}

type DeleteResult struct {
	GenericResult
	DelCount int64
	Err      error
}

type CountResult struct {
	GenericResult
	Count int64
	Err   error
}

type CommandResult struct {
	GenericResult
	DbRes *mongo.SingleResult
	Err   error
}

type IndexCreateResult struct {
	GenericResult
	IndexName string
	Err       error
}

type IndexDropResult struct {
	GenericResult
	Doc bson.Raw
	Err error
}

type IndexListResult struct {
	GenericResult
	Result []interface{}
	Err    error
}
