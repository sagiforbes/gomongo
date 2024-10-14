package main

import (
	"fmt"
	"time"

	"github.com/sagiforbes/gomongo"
	"go.mongodb.org/mongo-driver/bson"
)

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

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func InsertOne(gmc *gomongo.Client) error {
	user_doc := User{
		"1004", "name 4", "last 4", "999=1111111",
	}

	result := gomongo.InsertOneSync(gmc, "user", user_doc)
	fmt.Println("insert result", result.DbRes)
	return result.Err
}

func InsertMany(gmc *gomongo.Client) error {
	user_docs := []User{
		{"1000", "name 1", "last 1", "222=3333333"},
		{"1001", "name 2", "last 2", "555=6666666"},
	}

	user_addr := []Address{
		{"1000", "my address 1"},
		{"1001", "my address 2"},
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

func main() {
	mongo_host := "mongodb://127.0.0.1"
	gmc := gomongo.NewClient(mongo_host, "my_database", 60*time.Second)
	err := InsertMany(gmc)
	panicOnError(err)

	err = InsertOne(gmc)
	panicOnError(err)

	err = findUserAddress(gmc, "1000")
	panicOnError(err)

}
