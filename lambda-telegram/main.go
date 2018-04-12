package main

import (
	"os"

	"github.com/LaurenceUsas/car-ad-helper/carbot"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	// Dynamo DB
	dbRegion     = "eu-west-1"
	dbTableName  = "car-ad-helper"
	dbPrimaryKey = "userID"
)

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	cb := carbot.NewCarbot(
		os.Getenv("BOT_TOKEN"),
		dbRegion,
		dbTableName,
		dbPrimaryKey,
		os.Getenv("REG_PASS"),
		os.Getenv("URL_SCRAPER"),
	)

	var err error
	if request.Body == "" {
		err = cb.CheckAll() //os.Getenv("ALL_PASS")
	} else {
		err = cb.Respond(request.Body)
	}

	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 200}, nil
	}
	return events.APIGatewayProxyResponse{Body: "OK", StatusCode: 200}, nil
}
