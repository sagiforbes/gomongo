# gomongo
Gomongo is a very thin wrapper to mongo go driver. If increase productivity and code reliablity by using generics to represent a document. Also this wrapper respond with a single structure that holds all the inforamtion you need. 

Each mongo function can be executed as sync or go function. In case you are using go function, this module returns a none blocking channel. You can drain the channel whenever you like to get the result of the operation. This is good for starting several actions and then collect the result down the code.

# Installing
use the following command to add gomongo to your project
```bash

go get github.com/sagiforbes/gomongo


```


# Usage examples


## create client struct

You start by creating a client struct to be used by all gomongo functions. In this example we run mongo as a single node on localhost. To get a client run

```go
import "github.com/sagiforbes/gomongo"

func main(){
  mongo_host="mongodb://127.0.0.1"
  var gmc = gomongo.NewClient(mongo_host, "my_database", 60*time.Second)
}

```

`gmc` can be used with all gomongo functions. 
`my_database` is the name of the database you want to use.
`60*time.Seconds` the timeout of all mongo related functions.

## Data structurs

Throughout these examples we use two structs: user, address. Their definitions are:

```go
type User struct {
	Id     string `bson:"id" json:"id"`
	Name   string `bson:"name" json:"name"`
	Last   string `bson:"last" json:"last"`
	Mobile string `bson:"mobile" json:"mobile"`
}

type Address struct {
	UserId string `bson:"userId" json:"userId"`
	Addr   string `bson:"addr" json:"addr"`
}

```



## Insert new records

Follow are example to async insertion of records to the database

```go

func InsertMany(gmc *gomongo.Client) error {
	user_docs := []User{
		{"1000", "name 1", "last 1", "222=3333333"},
		{"1001", "name 2", "last 2", "555=6666666"},
	}

	user_addr := []Address{
		{"1000", "my address 1"},
		{"1002", "my address 2"},
	}

	result := gomongo.InsertManySync(gmc, "users", user_docs)
	if result.Err != nil {
		return result.Err
	}
	fmt.Println("insert result", result.DbRes)

	result = gomongo.InsertManySync(gmc, "address", user_addr)
	panicOnError(result.Err)
	return result.Err
}


```

`gmc` was created in previouse example and passed as a parameter. Here we insert two documents to two different collections

As you can see we use the sync version of the methods: (InsertManySync).

If you wanted to start inserting the documents in parallel and check the result at the end, we could use the go version function, like this:

```go

func InsertMany(gmc *gomongo.Client) error {
	user_docs := []User{
		{"1000", "name 1", "last 1", "222=3333333"},
		{"1001", "name 2", "last 2", "555=6666666"},
	}

	user_addr := []Address{
		{"1000", "my address 1"},
		{"1002", "my address 2"},
	}

	insert_user_ch := gomongo.InsertMany(gmc, "users", user_docs)
	
	insert_addr_ch := gomongo.InsertMany(gmc, "address", user_addr)
	
  insert_user_res := <-insert_user_ch
  insert_addr_res := <-insert_addr_ch

  if insert_user_res.Err!=nil{
    fmt.Errorf("failed to insert to users collection %v",insert_user_res.Err)
  }

  if insert_addr_res.Err!=nil{
    fmt.Errorf("failed to insert to users collection %v",insert_addr_res.Err)
  }

}


```

Note that in this example we use the go routing version of `InsertMany` (no sync endien). 

This means that we first insert to `users` collection and immediatly insert docuemnts to `address` collections.

Only after we inserted all the documents we check the result of the operation. that is done by the opperation:

```go

insert_user_res:=<-insert_user_ch
insert_addr_res:=<-insert_addr_ch

```

Here we drain the channels and check the result of the operation

## Search for document

The advatage of using the go routing version is more apparent when we search for documents. 

Say we want to fetch a `user` along with his `address`. we need to make two calls to mongo. 

first call to fetch the user info and another to fetch the address. 
We can do this in parallel by using the go function of findOne

```go
import 	"go.mongodb.org/mongo-driver/bson"


func findUserAddress(gmc *gomongo.Client, userId string) error {
	user_ch := gomongo.FindOne[User](gmc, "users", bson.M{"id": userId})
	addr_ch := gomongo.FindOne[Address](gmc, "address", bson.M{"userId": userId})

	user_res := <-user_ch
	addr_res := <-addr_ch

	if user_res.Err != nil || addr_res.Err != nil {
		return fmt.Errorf("failed to fetch user or address info %v %v", user_res.Err, addr_res.Err)
	}

	if !user_res.Found {
		return fmt.Errorf("user %s, not found", userId)
	}

	if !addr_res.Found {
		return fmt.Errorf("address of user %s, not found", userId)
	}

	fmt.Printf("Found user %v and address %v\n", user_res.Document, addr_res.Document.Addr)

	return nil
}


```

Note that bson.M is defined at `go.mongodb.org/mongo-driver/bson` module.

In this example we make query to users and address in parallel. than we drain the results from the channels. We than check for errors and print the result. 

Another cool feature of generic, is that you have immidiat context of the document. Is is very helpful in large scale projects. 


## Other method

The module implement a wrapper to all major functions of mongodb, namely:

- InsertMany
- InsertOne
- UpdateOne
- UpdateMany
- BulkWrite
- ReplaceOne
- FindOne
- Find
- Distinct
- FindStream
- DeleteOne
- DeleteMany
- CountDocuments
- RunCommand
- CreateIndex
- DropIndex
- DropAllIndex


All these method in gomongo return a none blocking channel to the struct result. 
Each method of gomongo has a sync method that return the struct immediatly after execution. Namely:

- InsertManySync
- InsertOneSync
- UpdateOneSync
- UpdateManySync
- BulkWriteSync
- ReplaceOneSync
- FindOneSync
- FindSync
- DistinctSync
- FindStreamSync
- DeleteOneSync
- DeleteManySync
- CountDocumentsSync
- RunCommandSync
- CreateIndexSync
- DropIndexSync
- DropAllIndexSync