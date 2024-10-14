package main

import (
	"fmt"
	"time"

	"github.com/sagiforbes/gomongo"
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
		{"1002", "my address 2"},
	}

	result := gomongo.InsertManySync(gmc, "users", user_docs)
	if result.Err != nil {
		return result.Err
	}
	fmt.Println("insert result", result.DbRes)

	result = gomongo.InsertManySync(gmc, "address", user_addr)
	if result.Err != nil {
		return result.Err
	}
	return nil
}

func main() {
	mongo_host := "mongodb://127.0.0.1"
	gmc := gomongo.NewClient(mongo_host, "my_database", 60*time.Second)
	err := InsertMany(gmc)
	if err != nil {
		panic(err)
	}
	err = InsertOne(gmc)
	if err != nil {
		panic(err)
	}
}
