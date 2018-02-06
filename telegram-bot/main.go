package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type MyEvent struct {
	Name string `json:"What is your name?"`
	Age  int    `json:"How old are you?"`
}

type MyResponse struct {
	Message string `json:"Answer:"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var event MyEvent
	err := json.Unmarshal([]byte(request.Body), &event)
	if err != nil {
		fmt.Println("error:", err)
	}

	resp := MyResponse{Message: fmt.Sprintf("%s is %d years old!", event.Name, event.Age)}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("error:", err)
	}

	return events.APIGatewayProxyResponse{Body: string(b), StatusCode: 200}, nil
}

func main() {
	//lambda.Start(Handler)
	fmt.Println("Starting")
	SetWebhook("https://03kemj3d9b.execute-api.eu-west-1.amazonaws.com/dev/telegram-bot", "477956583:AAFqmt78SXk4OrSaGQSQ110dPNopun816EE")
	fmt.Println("Finished")
}
