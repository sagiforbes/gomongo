package gomongo

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Course struct {
	Name  string `bson:"name"`
	Score int32  `bson:"score"`
}

type Restaurant struct {
	Name         string
	RestaurantId string      `bson:"restaurant_id,omitempty"`
	Cuisine      string      `bson:"cuisine,omitempty"`
	Address      interface{} `bson:"address,omitempty"`
	Courses      []Course    `bson:"courses,omitempty"`
}

const HOST = "mongodb://localhost:27017"
const DB_NAME = "test_client"
const COLL_NAME_RESTAURANT = "restorant"
const COLL_NAME_COURSE = "course"

var client = NewClient(HOST, DB_NAME, time.Second*60)

func testClient(t *testing.T) {
	var _, err = client.GetMongoClient()
	if err != nil {
		t.Fatalf("Failed to connect to  %s", err.Error())
		t.FailNow()
	}
}

func testPing(t *testing.T) {
	if !client.Ping() {
		t.Fatalf("failed to ping")
		t.FailNow()
	}
}

func TestGomongoConnection(t *testing.T) {

	t.Run("connection", testClient)
	t.Run("ping", testPing)

}

/*
**********************************************************************************

**********************************************************************************
 */
func testInsertMany(t *testing.T) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 2", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 3", Cuisine: "cuise 3"},
	}

	res := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if res.Err != nil {
		t.Errorf("%s", res.Err)
	} else {
		t.Logf("insert many result %s", res.DbRes)
	}

	idx_res := <-CreateIndex(client, COLL_NAME_RESTAURANT, bson.M{"name": 1}, nil)
	if idx_res.Err != nil {
		t.Errorf("failed to create index on name field %s", idx_res.Err.Error())
	} else {
		t.Log("create index: ", idx_res.IndexName)
	}

}

func testInsertOne(t *testing.T) {
	newRestaurant := Restaurant{Name: "restorant single", Cuisine: "cuise single"}

	res := InsertOneSync(client, COLL_NAME_RESTAURANT, newRestaurant)
	if res.Err != nil {
		t.Errorf("%s, %s", res.Err, errors.Unwrap(res.Err))
	} else {
		t.Logf("insert many result %s", res.DbRes)
	}
}

func testFindOne(t *testing.T) {
	res := FindOneSync[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"})
	if res.Err != nil {
		t.Errorf("failed to find document %s, %s", res.Err, errors.Unwrap(res.Err))
	} else {
		if res.Found {
			t.Logf("found one restorant: %v", res.Document)

		} else {
			t.Errorf("did not find expecting restorant")
		}

	}
}

func testFindMany(t *testing.T) {
	res := FindSync[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"})
	if res.Err != nil {
		t.Errorf("failed to find document %s, %s", res.Err, errors.Unwrap(res.Err))
	} else {
		if len(res.Documents) > 0 {
			t.Logf("restorants: %v", res.Documents)

		} else {
			t.Errorf("did not find expecting restorant")
		}
	}

	t.Log("now checking async version")
	chanRes := <-Find[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"})
	if chanRes.Err != nil {
		t.Errorf("failed to find document %s, %s", chanRes.Err, errors.Unwrap(chanRes.Err))
	} else {
		if len(chanRes.Documents) > 0 {
			t.Logf("restorants: %v", chanRes.Documents)
		} else {
			t.Errorf("did not find expecting restorant")
		}
	}

}

func testFindStream(t *testing.T) {
	const totalDocs = 1000
	var data = make([]interface{}, totalDocs)
	for i := 0; i < totalDocs; i++ {
		courses := []Course{{Name: fmt.Sprintf("course name %d", i), Score: int32(i * 10)}}

		data[i] = Restaurant{
			Name:         "stream name",
			Cuisine:      fmt.Sprintf("stream cuisine %d", i),
			Address:      fmt.Sprintf("stream addres %d", i),
			RestaurantId: fmt.Sprintf("id %d", i),
			Courses:      courses}
	}

	res := InsertManySync(client, COLL_NAME_RESTAURANT, data)
	if res.Err != nil {
		t.Error("failed to insert documents", res.Err)
	}

	find := <-FindStream[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "stream name"})

	if find.Err != nil {
		t.Error("failed to find via stream ", find.Err)
	}

	cnt := 0
	for foundDocRes := range find.DocumentStream {
		if foundDocRes.Err != nil {
			t.Errorf("error reading doc %d", cnt)
			t.Error(foundDocRes.Err)
		}
		t.Logf("doc %d) %v", cnt, foundDocRes.Document)
		cnt++
	}
	if cnt < totalDocs {
		t.Error("not all documents received")
	}

}

func testDistinct(t *testing.T) {
	res := DistinctSync[string](client, COLL_NAME_RESTAURANT, "name", bson.M{})
	if res.Err != nil {
		t.Error(res.Err)
	}
	t.Log(res.Values)
}

func testNotFindOne(t *testing.T) {
	res := FindOneSync[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "not found restorant name"}, nil)
	if res.Err != nil {
		t.Errorf("failed to find document %s, %s", res.Err, errors.Unwrap(res.Err))
	} else {
		if res.Found {
			t.Errorf("found a record while should not have found")
		}

	}
}

func testDeleteOne(t *testing.T) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 3"},
	}

	resInst := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if resInst.Err != nil {
		t.Errorf("%s, %s", resInst.Err, errors.Unwrap(resInst.Err))
	}

	res := DeleteOneSync(client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"}, nil)

	if res.Err != nil {
		t.Errorf("failed to delete recipe %s", res.Err)
	}

	if res.DelCount < 1 {
		t.Errorf("no document deleted")
	} else {
		t.Logf("deleted %d documents", res.DelCount)
	}

}

func testDeleteMany(t *testing.T) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 3"},
	}

	resInst := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if resInst.Err != nil {
		t.Errorf("%s, %s", resInst.Err, errors.Unwrap(resInst.Err))
	}
	res := DeleteManySync(client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"}, nil)
	if res.Err != nil {
		t.Errorf("failed to delete recipe %s", res.Err)
	}

	if res.DelCount < 1 {
		t.Errorf("no document deleted")
	} else {
		t.Logf("deleted %d documents", res.DelCount)
	}

}

func testUpdateDocument(t *testing.T) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 1", Cuisine: "cuise 3"},
	}
	resInst := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if resInst.Err != nil {
		t.Errorf("%s", resInst.Err)
	}
	t.Log("updating restorate 1. All cuise will be called sagi 3")
	resUpdate := UpdateManySync(client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1"}, bson.M{"$set": bson.M{"cuisine": "sagi 3"}})
	if resUpdate.Err != nil {
		t.Errorf("failed at update many %s", resUpdate.Err)
	}

	resFind := FindSync[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"cuisine": "sagi 3"})
	if len(resFind.Documents) < 1 {
		t.Errorf("No document was updated at update check")
	}

	resFindNone := FindSync[Restaurant](client, COLL_NAME_RESTAURANT, bson.M{"name": "restorant 1", "cuisine": "cuise 1"})
	if len(resFindNone.Documents) != 0 {
		t.Errorf("Some document where not updated")
	}

}

func testBulkWrite(t *testing.T) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 2", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 3", Cuisine: "cuise 3"},
	}
	resInst := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if resInst.Err != nil {
		t.Errorf("%s", resInst.Err)
	}

	var new_values [3]string
	for i := range new_values {
		new_values[i] = fmt.Sprintf("cuise change by update %d", time.Now().UnixMilli())
	}
	var update_models []mongo.WriteModel

	for i := range new_values {
		uom := mongo.NewUpdateOneModel().SetFilter(bson.M{"name": fmt.Sprintf("restorant %d", i+1)}).SetUpdate(bson.M{"$set": bson.M{"cuisine": new_values[i]}})
		update_models = append(update_models, uom)
	}

	res := BulkWriteSync(client, COLL_NAME_RESTAURANT, update_models)
	if res.Err != nil {
		t.Error(res.Err)
	}
	if res.DbRes.ModifiedCount != int64(len(new_values)) {
		t.Errorf("updated %d while expected %d", res.DbRes.ModifiedCount, len(new_values))
	}

}

func TestGomongoWrite(t *testing.T) {
	t.Run("insert many", testInsertMany)

	t.Run("insert one", testInsertOne)

	t.Run("update", testUpdateDocument)

	t.Run("bulk write", testBulkWrite)
}

func TestGomongoRead(t *testing.T) {
	t.Run("find one", testFindOne)

	t.Run("do not find one", testNotFindOne)

	t.Run("find many", testFindMany)

	t.Run("find stream", testFindStream)

	t.Run("distinct", testDistinct)
}

func TestGomongoDelete(t *testing.T) {
	t.Run("delete one", testDeleteOne)

	t.Run("delete many", testDeleteMany)

}

/*******************************************************************************************




*******************************************************************************************/

func bmInsertManySync(b *testing.B) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 2", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 3", Cuisine: "cuise 3"},
	}

	b.Log("Sync operaions")
	b.ResetTimer()
	res := InsertManySync(client, COLL_NAME_RESTAURANT, newRestaurants)
	if res.Err != nil {
		b.Errorf("%s", res.Err)
	} else {
		b.Logf("insert many result %s", res.DbRes)
	}

}

func bmInsertManyAsync(b *testing.B) {
	newRestaurants := []interface{}{
		Restaurant{Name: "restorant 1", Cuisine: "cuise 1"},
		Restaurant{Name: "restorant 2", Cuisine: "cuise 2"},
		Restaurant{Name: "restorant 3", Cuisine: "cuise 3"},
	}
	b.Log("Async operaions")
	b.ResetTimer()
	resch := <-InsertMany(client, COLL_NAME_RESTAURANT, newRestaurants)
	if resch.Err != nil {
		b.Errorf("failed to insert many")
	}

	b.Log("Done waiting for all async")
}

func BenchmarkGomongoWrite(b *testing.B) {
	b.Run("insert many sync", bmInsertManySync)
	b.Run("insert many async", bmInsertManyAsync)
}
