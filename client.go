package gomo

import (
	"context"
	"log"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	host              string
	database          string
	connectionTimeout time.Duration
}

func (c *Client) ctx() (context.Context, context.CancelFunc) {
	parent := context.Background()
	return context.WithTimeout(parent, c.connectionTimeout)
}

func (c *Client) coll(collectionName string) (*mongo.Collection, error) {
	conn, err := c.GetMongoClient()
	if err != nil {
		return nil, err
	}

	return conn.Database(c.database).Collection(collectionName), nil
}

func (c *Client) GetMongoClient() (*mongo.Client, error) {
	ctx, cancelFunc := c.ctx()
	defer cancelFunc()
	opt := options.Client().ApplyURI(c.host)
	return mongo.Connect(ctx, opt)
}

func (c *Client) Ping() bool {
	client, err := c.GetMongoClient()
	if err != nil {
		return false
	}
	defer func() {
		if client != nil {
			client.Disconnect(context.TODO())
		}
	}()
	ctx, cncl := c.ctx()
	defer cncl()
	return client.Ping(ctx, nil) == nil
}

/*
****************************************************************************************************************
****************************************************************************************************************
****************************************************************************************************************

	SYNC methods

****************************************************************************************************************
****************************************************************************************************************
******************************************************************************************************************
*/
func InsertManySync(c *Client, collName string, documents []interface{}, opts ...*options.InsertManyOptions) WriteManyResult {
	coll, err := c.coll(collName)
	if err != nil {
		return WriteManyResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()

	insertRes, err := coll.InsertMany(ctx, documents, opts...)
	if err != nil {
		return WriteManyResult{Err: NewError(MsgGomongoInsertManyError, err), DbRes: *insertRes}
	}
	return WriteManyResult{Err: nil, DbRes: *insertRes}
}

// InsertOneSync insert one document to collection.
func InsertOneSync(c *Client, collName string, document interface{}, opts ...*options.InsertOneOptions) WriteOneResult {
	coll, err := c.coll(collName)
	if err != nil {
		return WriteOneResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()

	insertRes, err := coll.InsertOne(ctx, document, opts...)
	if err != nil {
		return WriteOneResult{Err: NewError(MsgGomongoInsertManyError, err), DbRes: nil}
	}
	return WriteOneResult{Err: nil, DbRes: insertRes}
}

func UpdateOneSync(c *Client, collName string, filter interface{}, instruction interface{}, opts ...*options.UpdateOptions) UpdateResult {
	coll, err := c.coll(collName)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()

	dbUpdateRes, err := coll.UpdateOne(ctx, filter, instruction, opts...)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoInsertManyError, err)}
	}

	return UpdateResult{Err: nil, DbRes: dbUpdateRes}
}

func UpdateManySync(c *Client, collName string, filter interface{}, instruction interface{}, opts ...*options.UpdateOptions) UpdateResult {
	coll, err := c.coll(collName)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()

	dbUpdateRes, err := coll.UpdateMany(ctx, filter, instruction, opts...)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoInsertManyError, err)}
	}

	return UpdateResult{Err: nil, DbRes: dbUpdateRes}
}

func ReplaceOneSync(c *Client, collName string, filter interface{}, document interface{}, opts ...*options.ReplaceOptions) UpdateResult {
	coll, err := c.coll(collName)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()

	dbUpdateRes, err := coll.ReplaceOne(ctx, filter, document, opts...)
	if err != nil {
		return UpdateResult{Err: NewError(MsgGomongoInsertManyError, err)}
	}

	return UpdateResult{Err: nil, DbRes: dbUpdateRes}
}

// FindOneSync sync version of searching for a single document in a collection
func FindOneSync[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOneOptions) ReadOneResult[T] {
	coll, err := c.coll(collName)
	if err != nil {
		return ReadOneResult[T]{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	singleRes := coll.FindOne(ctx, filter, opts...)
	if singleRes.Err() != nil {
		if singleRes.Err() == mongo.ErrNoDocuments {
			return ReadOneResult[T]{Found: false, Err: nil}
		}
		return ReadOneResult[T]{Err: NewError(MsgGomongoFailedFindError, singleRes.Err())}
	}

	var data T
	err = singleRes.Decode(&data)
	if err != nil {
		return ReadOneResult[T]{Err: NewError(MsgGomongoUnmarshalError, singleRes.Err())}
	}
	return ReadOneResult[T]{Document: data, Found: true, DbRes: singleRes}
}

// FindSync query for documents in a sync way
func FindSync[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOptions) ReadManyResult[T] {
	coll, err := c.coll(collName)
	if err != nil {
		return ReadManyResult[T]{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	cursor, cur_err := coll.Find(ctx, filter, opts...)
	if cur_err != nil {
		return ReadManyResult[T]{Err: NewError(MsgGomongoCursorError, cur_err)}
	}
	defer cursor.Close(context.TODO())
	var resultDocs []T

	fetchCtx, fetchCancelFunc := c.ctx()
	defer fetchCancelFunc()

	err = cursor.All(fetchCtx, &resultDocs)

	if err != nil {
		return ReadManyResult[T]{Documents: nil, Err: NewError(MsgGomongoFetchError, err)}
	}

	return ReadManyResult[T]{Documents: resultDocs, Err: nil}

}

func FindStreamSync[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOptions) ReadStreamResult[T] {
	coll, err := c.coll(collName)
	if err != nil {
		return ReadStreamResult[T]{DocumentStream: nil, Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	cursor, cur_err := coll.Find(ctx, filter, opts...)
	if cur_err != nil {
		return ReadStreamResult[T]{DocumentStream: nil, Err: NewError(MsgGomongoCursorError, cur_err)}
	}

	channel_buffer_size := 200
	if len(opts) > 1 {
		opt := opts[0]
		if opt.BatchSize != nil {
			channel_buffer_size = int(math.Max(200, float64(*opt.BatchSize*2)))
		}

	}
	log.Default().Printf("at find stream, using channel batch size %d\n", channel_buffer_size)
	docCh := make(chan ReadOneResult[T], channel_buffer_size)
	ret := ReadStreamResult[T]{DocumentStream: docCh}
	go func() {
		defer cursor.Close(context.TODO())
		defer close(docCh)
		for cursor.Next(context.TODO()) {
			var fetchedDoc T
			parseErr := cursor.Decode(&fetchedDoc)
			if parseErr != nil {
				docCh <- ReadOneResult[T]{Found: parseErr == nil, Err: NewError(MsgGomongoFailedFindError, parseErr)}
			} else {
				docCh <- ReadOneResult[T]{Found: parseErr == nil, Document: fetchedDoc, Err: nil}
			}

		}

		if cursor.Err() != nil {
			docCh <- ReadOneResult[T]{Found: false, Err: NewError(MsgGomongoFetchError, cursor.Err())}
		}

	}()

	return ret
}

// DeleteOneSync delete on document base on the filter string
// for options and details mongo deleteon command see: [https://www.mongodb.com/docs/drivers/go/current/usage-examples/deleteOne/]
func DeleteOneSync(c *Client, collName string, filter interface{}, opts ...*options.DeleteOptions) DeleteResult {
	coll, err := c.coll(collName)
	if err != nil {
		return DeleteResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	delRes, err := coll.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return DeleteResult{Err: NewError(MsgGomongoDeleteError, err)}
	}
	return DeleteResult{Err: nil, DelCount: delRes.DeletedCount}
}

// DeleteManySync delete many documents from a collection. work in a sync way
// for options and details mongo deleteon command see: [https://www.mongodb.com/docs/drivers/go/current/usage-examples/deleteOne/]
func DeleteManySync(c *Client, collName string, filter interface{}, opts ...*options.DeleteOptions) DeleteResult {
	coll, err := c.coll(collName)
	if err != nil {
		return DeleteResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	delRes, err := coll.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return DeleteResult{Err: NewError(MsgGomongoDeleteError, err)}
	}
	return DeleteResult{Err: nil, DelCount: delRes.DeletedCount}
}

// CountDocuments  count the documents that return from the filter
func CountDocumentsSync(c *Client, collName string, filter interface{}, opts ...*options.CountOptions) CountResult {
	coll, err := c.coll(collName)
	if err != nil {
		return CountResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	count, err := coll.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return CountResult{Err: NewError(MsgGomongoConnectionError, err)}
	}
	return CountResult{Count: count}
}

// RunCommand run a command on the database
func RunCommandSync(c *Client, cmd interface{}, opts ...*options.RunCmdOptions) CommandResult {
	conn, err := c.GetMongoClient()
	if err != nil {
		return CommandResult{Err: NewError(MsgGomongoConnectionError, err)}
	}
	ctx, cancel := c.ctx()
	defer cancel()
	singleRes := conn.Database(c.database).RunCommand(ctx, cmd, opts...)
	if singleRes.Err() != nil {
		return CommandResult{Err: NewError(MsgGomongoCommandError, singleRes.Err())}
	}
	return CommandResult{DbRes: singleRes}
}

func CreateIndexSync(c *Client, collName string, indexDef interface{}, idxOpt *options.IndexOptions) IndexCreateResult {
	coll, err := c.coll(collName)
	if err != nil {
		return IndexCreateResult{Err: NewError(MsgGomongoConnectionError, err)}
	}
	ctx, cancel := c.ctx()

	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    indexDef,
		Options: idxOpt,
	}

	name, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return IndexCreateResult{Err: NewError(MsgGomongoIndexError, err)}
	}
	return IndexCreateResult{IndexName: name}
}

func DropIndexSync(c *Client, collName string, name string, opts ...*options.DropIndexesOptions) IndexDropResult {
	coll, err := c.coll(collName)
	if err != nil {
		return IndexDropResult{Err: NewError(MsgGomongoConnectionError, err)}
	}
	ctx, cancel := c.ctx()
	defer cancel()
	raw, err := coll.Indexes().DropOne(ctx, name, opts...)
	if err != nil {
		return IndexDropResult{Err: NewError(MsgGomongoIndexError, err)}
	}
	return IndexDropResult{Doc: raw}
}

func DropAllIndexSync(c *Client, collName string, opts ...*options.DropIndexesOptions) IndexDropResult {
	coll, err := c.coll(collName)
	if err != nil {
		return IndexDropResult{Err: NewError(MsgGomongoConnectionError, err)}
	}
	ctx, cancel := c.ctx()
	defer cancel()
	raw, err := coll.Indexes().DropAll(ctx, opts...)
	if err != nil {
		return IndexDropResult{Err: NewError(MsgGomongoIndexError, err)}
	}
	return IndexDropResult{Doc: raw}
}

func ListIndexSync(c *Client, collName string, opts ...*options.ListIndexesOptions) IndexListResult {
	coll, err := c.coll(collName)
	if err != nil {
		return IndexListResult{Err: NewError(MsgGomongoConnectionError, err)}
	}

	ctx, cancel := c.ctx()
	defer cancel()
	cursor, cur_err := coll.Indexes().List(ctx, opts...)
	if cur_err != nil {
		return IndexListResult{Err: NewError(MsgGomongoCursorError, cur_err)}
	}
	defer cursor.Close(context.TODO())
	curs_ctx, curs_cancel := c.ctx()
	defer curs_cancel()

	var result []interface{}
	all_err := cursor.All(curs_ctx, &result)
	if all_err != nil {
		return IndexListResult{Err: NewError(MsgGomongoFetchError, all_err)}
	}
	return IndexListResult{Result: result}
}

/*
****************************************************************************************************************
****************************************************************************************************************
****************************************************************************************************************

	async methods

****************************************************************************************************************
****************************************************************************************************************
******************************************************************************************************************
*/

// InsertMany insert many document in async way
func InsertMany(c *Client, collName string, documents []interface{}, opts ...*options.InsertManyOptions) chan WriteManyResult {
	ret := make(chan WriteManyResult, 1)
	go func() {
		ret <- InsertManySync(c, collName, documents, opts...)
		close(ret)
	}()
	return ret
}

func InsertOne(c *Client, collName string, document interface{}, opts ...*options.InsertOneOptions) chan WriteOneResult {
	ret := make(chan WriteOneResult, 1)
	go func() {
		ret <- InsertOneSync(c, collName, document, opts...)
		close(ret)
	}()
	return ret
}

// Update update a single document in an async way
func UpdateOne(c *Client, collName string, filter interface{}, instruction interface{}, opts ...*options.UpdateOptions) chan UpdateResult {
	ret := make(chan UpdateResult, 1)
	go func() {
		ret <- UpdateOneSync(c, collName, filter, instruction, opts...)
		close(ret)
	}()
	return ret
}

// Update update all documents matching the filter creteria in an async way
func UpdateMany(c *Client, collName string, filter interface{}, instruction interface{}, opts ...*options.UpdateOptions) chan UpdateResult {
	ret := make(chan UpdateResult, 1)
	go func() {
		ret <- UpdateManySync(c, collName, filter, instruction, opts...)
		close(ret)
	}()
	return ret
}

// Update update a single document in an async way
func ReplaceOne(c *Client, collName string, filter interface{}, document interface{}, opts ...*options.ReplaceOptions) chan UpdateResult {
	ret := make(chan UpdateResult, 1)
	go func() {
		ret <- ReplaceOneSync(c, collName, filter, document, opts...)
		close(ret)
	}()
	return ret
}

// FindOne async function that search for a single document in a collection
func FindOne[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOneOptions) chan ReadOneResult[T] {
	ret := make(chan ReadOneResult[T], 1)
	go func() {
		ret <- FindOneSync[T](c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// Find query for documents in async way
func Find[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOptions) chan ReadManyResult[T] {
	ret := make(chan ReadManyResult[T], 1)
	go func() {
		ret <- FindSync[T](c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// Find query for documents in async way
func FindStream[T any](c *Client, collName string, filter interface{}, opts ...*options.FindOptions) chan ReadStreamResult[T] {
	ret := make(chan ReadStreamResult[T], 1)
	go func() {
		ret <- FindStreamSync[T](c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// DeleteOne delete on document in a async way
// for options and details mongo deleteon command see: [https://www.mongodb.com/docs/drivers/go/current/usage-examples/deleteOne/]
func DeleteOne(c *Client, collName string, filter interface{}, opts ...*options.DeleteOptions) chan DeleteResult {
	ret := make(chan DeleteResult, 1)
	go func() {
		ret <- DeleteOneSync(c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// DeleteMany delete many documents from a collection. work in async way. you need to check the result of the channel
// for options and details mongo deleteon command see: [https://www.mongodb.com/docs/drivers/go/current/usage-examples/deleteOne/]
func DeleteMany(c *Client, collName string, filter interface{}, opts ...*options.DeleteOptions) chan DeleteResult {
	ret := make(chan DeleteResult, 1)
	go func() {
		ret <- DeleteManySync(c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// CountDocuments async, count the documents that return from the filter
func CountDocuments(c *Client, collName string, filter interface{}, opts ...*options.CountOptions) chan CountResult {
	ret := make(chan CountResult, 1)
	go func() {
		ret <- CountDocumentsSync(c, collName, filter, opts...)
		close(ret)
	}()
	return ret
}

// RunCommand run database command async
func RunCommand(c *Client, cmd interface{}, opts ...*options.RunCmdOptions) chan CommandResult {
	ret := make(chan CommandResult, 1)
	go func() {
		ret <- RunCommandSync(c, cmd, opts...)
		close(ret)
	}()
	return ret
}

func CreateIndex(c *Client, collName string, indexDef interface{}, opt *options.IndexOptions) chan IndexCreateResult {
	ret := make(chan IndexCreateResult, 1)
	go func() {
		ret <- CreateIndexSync(c, collName, indexDef, opt)
		close(ret)
	}()
	return ret
}

func DropIndex(c *Client, collName string, indexName string, opts ...*options.DropIndexesOptions) chan IndexDropResult {
	ret := make(chan IndexDropResult, 1)
	go func() {
		ret <- DropIndexSync(c, collName, indexName, opts...)
		close(ret)
	}()
	return ret
}

func DropAllIndex(c *Client, collName string, opts ...*options.DropIndexesOptions) chan IndexDropResult {
	ret := make(chan IndexDropResult, 1)
	go func() {
		ret <- DropAllIndexSync(c, collName, opts...)
		close(ret)
	}()
	return ret
}

// JoinMany alow waiting for seceral channels. the result of this method are channels that needs to be drained
func JoinMany[T any](sourceChans ...chan T) []T {
	var ret = make([]T, len(sourceChans))
	for i := 0; i < len(sourceChans); i++ {
		ret[i] = <-sourceChans[i]
	}
	return ret
}

// NewClient return a pointer to gomongo.Client that is used to communicate the Mongodb
//
// Parameters:
//
//	host: a host name in the form of mongodb://host:port. Fool explanation can be found at [https://www.mongodb.com/docs/drivers/go/current/fundamentals/connections/connection-guide/]
//	database: The client is aim to work with a single database. so you can set its name at the initialization stage.
//	connTimeout: The timeout of the operations against the database. If you do not set it will set to the default of 10 seconds
func NewClient(host string, database string, connTimeout ...time.Duration) *Client {
	ret := &Client{
		host:              host,
		database:          database,
		connectionTimeout: time.Duration(time.Second * 10),
	}
	if len(connTimeout) > 0 {
		ret.connectionTimeout = connTimeout[0]
	}

	return ret
}
