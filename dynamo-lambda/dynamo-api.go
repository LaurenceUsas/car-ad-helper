package cahdynamo

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

//================================================
//				Dynamo DB User
//================================================

// Mapped to Dynamo DB database model.
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

// TODO Add Query Limits.
// Change to pass on message through Err?
func (dbu *DBUser) QueryAdd(new string) {
	if dbu.Queries == nil {
		dbu.Queries = []string{}
	}

	if !dbu.QueryExist(new) {
		dbu.Queries = append(dbu.Queries, new)
	}
}

// Delete User.Queries by string
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

// Delete User.Queries[ID]
func (dbu *DBUser) QueryDeleteID(input int) {
	if dbu.Queries == nil {
		return
	}
	dbu.Queries = append(dbu.Queries[:input], dbu.Queries[input+1:]...)
}

// QueryExist checks presence in User.Query
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

// Adds map to the Cars
func (dbu *DBUser) CarsAdd(new map[string]bool) {
	if dbu.Cars == nil {
		dbu.Cars = make(map[string]bool)
	}

	for k, v := range new {
		dbu.Cars[k] = v
	}
}

//================================================
//				Dynamo DB API
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

// Store uploads DBUser to DynamoDB.
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

	log.Printf("Successfully added UserID:%v to %v\n", item.ID, api.TableName)
	return nil
}

// Retrieve DBUser by ID
func (api *DynamoAPI) Retrieve(userID int64) (*DBUser, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)
	if err != nil {
		fmt.Println(1)
		return nil, err
	}

	svc := dynamodb.New(sess)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(api.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			api.PrimaryKey: {N: aws.String(strconv.Itoa(int(userID)))},
		},
	})
	if err != nil {
		fmt.Println(2)
		fmt.Println()
		return nil, err
	}

	user := DBUser{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		fmt.Println(3)
		return nil, err
	}

	//TODO Better check.
	if user.ID == 0 {
		log.Printf("Could not find user [%d]\n", userID)
		return nil, nil
	}
	log.Printf("Retrieve of user [%v] successful", userID)
	return &user, nil
}

func (api *DynamoAPI) RetrieveAll() ([]DBUser, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)
	if err != nil {
		return nil, err
	}

	svc := dynamodb.New(sess)
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(api.TableName),
	})
	if err != nil {
		return nil, err
	}

	out := []DBUser{}
	for _, v := range result.Items {
		user := DBUser{}
		err = dynamodbattribute.UnmarshalMap(v, &user)
		if err != nil {
			return nil, err
		}
		out = append(out, user)
	}
	log.Printf("Retrieve of all [%v] users successful", len(out))
	return out, nil
}

// Delete DBUser from DynamoDB
func (api *DynamoAPI) Delete(item DBUser) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(api.Region)},
	)
	if err != nil {
		return err
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(api.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			api.TableName: {N: aws.String(strconv.Itoa(int(item.ID)))},
		},
	}

	svc := dynamodb.New(sess)
	_, err = svc.DeleteItem(input)

	if err != nil {
		return err
	}

	log.Printf("User [%d] deleted successfully", item.ID)
	return nil
}
