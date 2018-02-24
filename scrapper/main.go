package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ScrapeRequest struct {
	Queries string
}

func NewScrapeRequest(query string) *ScrapeRequest { // Force using difference request for each query
	sr := &ScrapeRequest{
		Queries: query,
	}
	return sr
}

type ScrapeResponse struct {
	Results map[string]bool
}

func NewScrapeResponse(results map[string]bool) *ScrapeResponse {
	sr := &ScrapeResponse{
		Results: results,
	}
	return sr
}

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//unmarshal to object.
	var req ScrapeRequest
	json.Unmarshal([]byte(request.Body), &req)

	l := getAutoplius(req.Queries)
	sr := NewScrapeResponse(l)
	srJson, _ := json.Marshal(sr)
	log.Printf("Returning %v ", srJson)

	return events.APIGatewayProxyResponse{Body: string(srJson), StatusCode: 200}, nil
}

func getAutoplius(address string) map[string]bool {
	list := make(map[string]bool, 2)

	url := strings.TrimPrefix(address, "https://autoplius.lt") // prep first. Storing with prefix to db for easier human lookup
	fmt.Println(url)
	for url != "" {
		fmt.Printf("Checking URL [%s]\n", url)
		doc, err := goquery.NewDocument("https://autoplius.lt" + url)
		if err != nil {
			log.Fatal(err)
		}

		//Links to the cars.
		doc.Find(".announcement-item").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")
			list[link] = true
		})

		// Link to next page.
		url = ""
		doc.Find(".next").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), "Kitas") {
				url, _ = s.Attr("href")
				if url != "" {
					return false
				}
			}
			return true
		})
	}
	return list
}
