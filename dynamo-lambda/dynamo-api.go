package cahdynamo

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	EmptyListValue = "Empty list"
)

//================================================
//
//================================================

type DBUser struct {
	ID            int64           `json:"userID"`
	Queries       []string        `json:"Queries"`
	Cars          map[string]bool `json:"Cars"`
	NotInterested map[string]bool `json:"NotInterested"`
	AutoUpdate    bool            `json:"AutoUpdate"`
}

func NewDBUser(id int64) *DBUser {
	u := &DBUser{
		ID:            id,
		Queries:       []string{},
		Cars:          map[string]bool{},
		NotInterested: map[string]bool{},
	}
	return u
}

// Add Query Limits.
// Change to pass on message through Err?
func (dbu *DBUser) QueryAdd(new string) {
	if dbu.Queries == nil {
		dbu.Queries = []string{}
	}

	if !dbu.QueryExist(new) {
		dbu.Queries = append(dbu.Queries, new)
	}
}

func (dbu *DBUser) QueryDeleteString(input string) {
	if dbu.Queries == nil {
		return
	}

	del := -1
	for i, v := range dbu.Queries {
		if v == input {
			del = i
			break
		}
	}
	if del != -1 {
		dbu.Queries = append(dbu.Queries[:del], dbu.Queries[del+1:]...)
	}
}

func (dbu *DBUser) QueryDeleteID(input int) {
	if dbu.Queries == nil {
		return
	}

	dbu.Queries = append(dbu.Queries[:input], dbu.Queries[input+1:]...)
}

// QueryExist checks if the input is present in User.Query
func (dbu *DBUser) QueryExist(input string) bool {
	if dbu.Queries == nil {
		return false
	}

	out := false
	for _, v := range dbu.Queries {
		if v == input {
			out = true
			break
		}
	}
	return out
}

func (dbu *DBUser) CarsAdd(new map[string]bool) {
	for k, v := range new {
		dbu.Cars[k] = v
	}
}

//================================================
//
//================================================

type DynamoAPI struct {
	Region     string
	TableName  string
	PrimaryKey string
}

func NewDynamoAPI(region, tableName, primaryKey string) *DynamoAPI {
	d := &DynamoAPI{
		Region:     region,
		TableName:  tableName,
		PrimaryKey: primaryKey,
	}
	return d
}

func (api *DynamoAPI) Store(item *DBUser) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)

	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(item)

	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(api.TableName),
	}

	if err != nil {
		return err
	}

	_, err = svc.PutItem(input)

	if err != nil {
		return err
	}

	fmt.Printf("Successfully added UserID:%v to %v\n", item.ID, api.TableName)
	return nil
}

// Retreive item data by primary key
func (api *DynamoAPI) Retrieve(userID int64) (*DBUser, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)

	if err != nil {
		fmt.Printf("%s", err)
	}

	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(api.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			api.TableName: {N: aws.String(strconv.Itoa(int(userID)))},
		},
	})

	user := DBUser{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)

	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	if user.ID == 0 {
		fmt.Printf("Could not find user [%d]\n", userID)
		return nil, nil
	}
	return nil, &user
}

func (api *DynamoAPI) RetrieveAll() ([]DBUser, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)

	if err != nil {
		fmt.Printf("%s", err)
	}

	svc := dynamodb.New(sess)

	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(api.TableName),
	})

	out := []DBUser{}
	for _, v := range result.Items {
		user := DBUser{}
		err = dynamodbattribute.UnmarshalMap(v, &user)
		if err != nil {
			panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		}
		out = append(out, user)
	}

	return nil, out
}

func (apie *DynamoAPI) Delete(item DBUser) {

}
