package cahdynamo

import (
	"log"
	"testing"
)

/*
var fibTests = []struct {
  n        int // input
  expected int // expected result
}{
  {1, 1},
  {2, 1},
  {3, 2},
  {4, 3},
  {5, 5},
  {6, 8},
  {7, 13},
}


*/

const (
	// Dynamo DB
	dbRegion     = "eu-west-1"
	dbTableName  = "car-ad-helper"
	dbPrimaryKey = "userID"
)

func TestQueryAdd(t *testing.T) {

}

func TestQueryDeleteString(t *testing.T) {

}

func TestQueryDeleteID(t *testing.T) {

}

func TestQueryExist(t *testing.T) {

}

func TestCarsAdd(t *testing.T) {

}

func TestStore(t *testing.T) {
	api := NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	u := NewDBUser(999)
	err := api.Store(u)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestRetrieve(t *testing.T) {

	api := NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	u, err := api.Retrieve(327840258)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	log.Println(u)
}

func TestRetrieveAll(t *testing.T) {

}

func TestDelete(t *testing.T) {

}
