package cahdynamo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -cover
// go test -coverprofile=name
// go tools cover -html=name

const (
	// Dynamo DB
	dbRegion     = "eu-west-1"
	dbTableName  = "car-ad-helper"
	dbPrimaryKey = "userID"
)

func TestDBUser(t *testing.T) {
	assert := assert.New(t)
	id := int64(999999999999)
	u := NewDBUser(id)

	assert.Equal(id, u.ID)
	assert.Len(u.Queries, 0)
	assert.Len(u.Cars, 0)
	assert.Len(u.NotInterested, 0)

	// Add
	a := "http://new.query.com/testing"
	b := a + "1"
	assert.Equal(false, u.QueryExist(a))
	u.QueryAdd(a)
	assert.Equal(true, u.QueryExist(a))
	u.QueryAdd(b)
	assert.Contains(u.Queries, a)
	assert.Len(u.Queries, 2)
	// Delete ID
	u.QueryDeleteString("asd")
	assert.Len(u.Queries, 2)
	u.QueryDeleteString(a)
	assert.NotContains(u.Queries, a)
	// Delete string
	u.QueryDeleteID(2)
	assert.Contains(u.Queries, b)
	u.QueryDeleteID(0)
	assert.Len(u.Queries, 0)
	// Cars Add
	cars := map[string]bool{
		"Car1": true,
		"Car2": false,
	}
	u.CarsAdd(cars)
	assert.Len(u.Cars, 2)
}

func TestIntegrationDynamoDB(t *testing.T) {
	assert := assert.New(t)
	db := NewDynamoAPI(dbRegion, dbTableName, dbPrimaryKey)
	id := int64(999999999999)
	u := NewDBUser(id)

	_, err := db.Retrieve(id)
	assert.NotEqual(nil, err)

	_, err = db.RetrieveAll()
	assert.Equal(nil, err)

	db.Store(u)
	db.Retrieve(id)
	assert.Equal(nil, err)

	err = db.Delete(u)
	assert.NotEqual(nil, err)
	db.Retrieve(id)
	assert.NotEqual(nil, err)

	db.Retrieve(id)
	db.RetrieveAll()
	// db.Store(u)
	db.Retrieve(id)
}

// func TestNewDBUser(t *testing.T) {
// 	type args struct {
// 		id int64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *DBUser
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewDBUser(tt.args.id); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewDBUser() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDBUser_QueryAdd(t *testing.T) {
// 	type fields struct {
// 		ID            int64
// 		Queries       []string
// 		Cars          map[string]bool
// 		NotInterested map[string]bool
// 		AutoUpdate    bool
// 	}
// 	type args struct {
// 		new string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 		{},
// 		{},
// 		{},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			dbu := &DBUser{
// 				ID:            tt.fields.ID,
// 				Queries:       tt.fields.Queries,
// 				Cars:          tt.fields.Cars,
// 				NotInterested: tt.fields.NotInterested,
// 				AutoUpdate:    tt.fields.AutoUpdate,
// 			}
// 			dbu.QueryAdd(tt.args.new)
// 		})
// 	}
// }
