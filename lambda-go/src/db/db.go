// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package db

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-gorp/gorp"
	"github.com/guregu/dynamo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
)

type DBConfig struct {
	DbService  *dynamodb.DynamoDB
	PrimaryKey string
	SortKey    string
	TableName  string
}

var DB *dynamo.DB
var WaterColorSiteTable dynamo.Table
var SessionTable dynamo.Table
var UserTable dynamo.Table
var sqlite *sql.DB
var dbmap *gorp.DbMap

func init() {
	if env.IsLocal {
		DB = dynamo.New(session.New(), &aws.Config{
			Region:   aws.String("ap-northeast-1"),
			Endpoint: aws.String("http://XXX.XXX.XXX.XXX:9000/"), //開発環境のDynamo
		})
		var err error
		sqlite, err = sql.Open("sqlite3", "../sqlite/db.sqlite")
		if err != nil {
			log.Fatal(err)
		}
		dbmap = &gorp.DbMap{Db: sqlite, Dialect: gorp.SqliteDialect{}}

	} else {
		DB = dynamo.New(session.New(), &aws.Config{
			Region: aws.String("ap-northeast-1"),
		})
		var err error
		sqlite, err = sql.Open("sqlite3", "./db.sqlite")
		if err != nil {
			log.Fatal(err)
		}
		dbmap = &gorp.DbMap{Db: sqlite, Dialect: gorp.SqliteDialect{}}
	}
	WaterColorSiteTable = DB.Table("wcs-table-prod")
	SessionTable = DB.Table("wcs-session-table-prod")
	UserTable = DB.Table("wcs-user-table-prod")
}

func GetPigment(lang model.SupportedLang, filtergroup int32, name string) (*[]model.Pigment, error) {
	var equipments []model.Pigment
	if !lang.IsSupportedLang() {
		return &equipments, nil
	}
	_, err := dbmap.Select(&equipments,
		"select key, name from pigments_"+lang.String()+
			" where filtergroup=? and name like ? order by case when name = ? then 1 else 2 end, key desc limit 10;", filtergroup, name+"%", name)
	if err != nil {
		log.Fatal(err)
		return &equipments, err
	}
	return &equipments, nil
}

func GetPainting(userId string, timestamp string) (model.Painting, error) {
	var result model.Painting
	err := WaterColorSiteTable.Get("UserId", userId).Range("Timestamp", dynamo.Equal, timestamp).One(&result)
	return result, err
}
func GetWaterColorSite(date string) (*[]model.Painting, error) {
	var result []model.Painting
	err := WaterColorSiteTable.Get("Date", date).Range("Timestamp", dynamo.Between, "", "").Index("wcs-table-prod-by-date").Limit(10).All(&result)
	return &result, err
}

func GetSession(sessionId string) (model.Session, error) {
	var result model.Session
	err := SessionTable.Get("SessionId", sessionId).One(&result)
	return result, err
}

func PutSession(session model.Session) (model.Session, error) {
	err := SessionTable.Put(session).Run()
	return session, err
}

func PutUser(user model.User) (model.User, error) {
	err := UserTable.Put(user).Run()
	return user, err
}

func GetUser(userId string) (model.User, error) {
	var result model.User
	err := UserTable.Get("UserId", userId).One(&result)
	return result, err
}

func PutPainting(painting *model.Painting) error {
	err := WaterColorSiteTable.Put(painting).Run()
	return err
}

//init setup teh session and define table name, primary key and sort key
func DBInit(tn string, pk string, sk string) DBConfig {

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	dbSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create DynamoDB client
	return DBConfig{
		DbService:  dynamodb.New(dbSession),
		PrimaryKey: pk,
		SortKey:    sk,
		TableName:  tn,
	}
}

func (dbc DBConfig) Save(prop interface{}) (interface{}, error) {
	av, err := dynamodbattribute.MarshalMap(prop)
	if err != nil {
		fmt.Println("Got error marshalling new property item:")
		fmt.Println(err.Error())
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
	}
	return prop, err
}

func (dbc DBConfig) Delete(prop interface{}) (interface{}, error) {
	av, err := dynamodbattribute.MarshalMap(prop)
	if err != nil {
		fmt.Println("Got error marshalling new property item:")
		fmt.Println(err.Error())
	}

	input := &dynamodb.DeleteItemInput{
		Key:       av,
		TableName: aws.String(dbc.TableName),
	}

	_, err = dbc.DbService.DeleteItem(input)
	if err != nil {
		fmt.Println("Got error calling DeetItem:")
		fmt.Println(err.Error())
	}
	return prop, err
}

//TODO: to evaluate th value of this tradeoff: this is probably a little slow but abstract the complexity for all uses of
//the save many function(and actually any core operation on array of interface)
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

//Writtes many items to a single table
func (dbc DBConfig) SaveMany(data interface{}) error {
	//Dynamo db currently limits batches to 25 items
	// batches := core.Chunk(InterfaceSlice(data), 25)
	// for i, dataArray := range batches {

	// 	log.Printf("DB> Batch %i inserting: %+v", i, dataArray)
	// 	items := make([]*dynamodb.WriteRequest, len(dataArray), len(dataArray))
	// 	for i, item := range dataArray {
	// 		av, err := dynamodbattribute.MarshalMap(item)
	// 		if err != nil {
	// 			fmt.Println("Got error marshalling new property item:")
	// 			fmt.Println(err.Error())
	// 		}
	// 		items[i] = &dynamodb.WriteRequest{
	// 			PutRequest: &dynamodb.PutRequest{
	// 				Item: av,
	// 			},
	// 		}
	// 	}

	// 	bwii := &dynamodb.BatchWriteItemInput{
	// 		RequestItems: map[string][]*dynamodb.WriteRequest{
	// 			dbc.TableName: items,
	// 		},
	// 	}

	// 	_, err := dbc.DbService.BatchWriteItem(bwii)
	// 	if err != nil {
	// 		if aerr, ok := err.(awserr.Error); ok {
	// 			switch aerr.Code() {
	// 			case dynamodb.ErrCodeProvisionedThroughputExceededException:
	// 				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
	// 			case dynamodb.ErrCodeResourceNotFoundException:
	// 				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
	// 			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
	// 				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
	// 			case dynamodb.ErrCodeRequestLimitExceeded:
	// 				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
	// 			case dynamodb.ErrCodeInternalServerError:
	// 				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
	// 			default:
	// 				fmt.Println(aerr.Error())
	// 			}
	// 		} else {
	// 			// Print the error, cast err to awserr.Error to get the Code and
	// 			// Message from an error.
	// 			fmt.Println(err.Error())
	// 		}
	// 		return err
	// 	}
	// }
	return nil
}

//Deletes many items to a single table
func (dbc DBConfig) DeleteMany(data interface{}) error {
	//Dynamo db currently limits batches to 25 items
	// batches := core.Chunk(InterfaceSlice(data), 25)
	// for i, dataArray := range batches {

	// 	log.Printf("DB> Batch %i deleting: %+v", i, dataArray)
	// 	items := make([]*dynamodb.WriteRequest, len(dataArray), len(dataArray))
	// 	for i, item := range dataArray {
	// 		av, err := dynamodbattribute.MarshalMap(item)
	// 		if err != nil {
	// 			fmt.Println("Got error marshalling new property item:")
	// 			fmt.Println(err.Error())
	// 		}
	// 		items[i] = &dynamodb.WriteRequest{
	// 			DeleteRequest: &dynamodb.DeleteRequest{
	// 				Key: av,
	// 			},
	// 		}
	// 	}

	// 	bwii := &dynamodb.BatchWriteItemInput{
	// 		RequestItems: map[string][]*dynamodb.WriteRequest{
	// 			dbc.TableName: items,
	// 		},
	// 	}

	// 	_, err := dbc.DbService.BatchWriteItem(bwii)
	// 	if err != nil {
	// 		if aerr, ok := err.(awserr.Error); ok {
	// 			switch aerr.Code() {
	// 			case dynamodb.ErrCodeProvisionedThroughputExceededException:
	// 				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
	// 			case dynamodb.ErrCodeResourceNotFoundException:
	// 				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
	// 			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
	// 				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
	// 			case dynamodb.ErrCodeRequestLimitExceeded:
	// 				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
	// 			case dynamodb.ErrCodeInternalServerError:
	// 				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
	// 			default:
	// 				fmt.Println(aerr.Error())
	// 			}
	// 		} else {
	// 			// Print the error, cast err to awserr.Error to get the Code and
	// 			// Message from an error.
	// 			fmt.Println(err.Error())
	// 		}
	// 		return err
	// 	}
	// }
	return nil
}

func (dbc DBConfig) Get(pk string, sk string, data interface{}) error {
	av := map[string]*dynamodb.AttributeValue{
		dbc.PrimaryKey: {
			S: aws.String(pk),
		},
	}
	if sk != "" {
		av[dbc.SortKey] = &dynamodb.AttributeValue{
			S: aws.String(sk),
		}
	}

	result, err := dbc.DbService.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(dbc.TableName),
		Key:       av,
	})
	if err != nil {
		fmt.Println("NOT FOUND")
		fmt.Println(err.Error())
		return err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, data)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return err
}

func (dbc DBConfig) FindStartingWith(pk string, value string, data interface{}) error {
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		KeyConditions: map[string]*dynamodb.Condition{
			dbc.PrimaryKey: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(pk),
					},
				},
			},
			dbc.SortKey: {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
		},
	}

	var result, err = dbc.DbService.Query(queryInput)
	if err != nil {
		fmt.Println("DB:FindStartingWith> NOT FOUND")
		fmt.Println(err.Error())
		return err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, data)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return err
}

func (dbc DBConfig) FindByGsi(value string, indexName string, indexPk string, data interface{}) error {
	var queryInput = &dynamodb.QueryInput{
		TableName: aws.String(dbc.TableName),
		IndexName: aws.String(indexName),
		KeyConditions: map[string]*dynamodb.Condition{
			indexPk: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
		},
	}

	var result, err = dbc.DbService.Query(queryInput)
	if err != nil {
		fmt.Println("NOT FOUND")
		fmt.Println(err.Error())
		return err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, data)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}
	return err
}
