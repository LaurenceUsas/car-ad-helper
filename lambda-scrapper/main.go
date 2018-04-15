package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/LaurenceUsas/car-ad-helper/carbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//unmarshal to object.
	var req carbot.ScrapeRequest
	json.Unmarshal([]byte(request.Body), &req)

	l, err := getAutoplius(req.Queries)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	sr := carbot.NewScrapeResponse(l)
	srJson, err := json.Marshal(sr)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	log.Printf("Returning %v ", srJson)

	return events.APIGatewayProxyResponse{Body: string(srJson), StatusCode: 200}, nil
}

func getAutoplius(address string) (map[string]bool, error) {
	const nextPageTag = "Kitas"

	list := make(map[string]bool, 2)

	url := strings.TrimPrefix(address, "https://autoplius.lt") // prep first. Storing with prefix to db for easier human lookup
	fmt.Println(url)
	for url != "" {
		fmt.Printf("Checking URL [%s]\n", url)
		doc, err := goquery.NewDocument("https://autoplius.lt" + url)
		if err != nil {
			return nil, err
		}

		//Links to the cars.
		doc.Find(".announcement-item").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")
			list[link] = true
		})

		// Link to next page.
		url = ""
		doc.Find(".next").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), nextPageTag) {
				url, _ = s.Attr("href")
				if url != "" {
					return false
				}
			}
			return true
		})
	}
	return list, nil
}
