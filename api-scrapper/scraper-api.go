package scrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

type ScrapperAPI struct {
	endpoint string
}

func NewScrapperAPI(link string) *ScrapperAPI {
	d := &ScrapperAPI{
		endpoint: link,
	}
	return d
}

func (api *ScrapperAPI) Invoke(url string) *ScrapeResponse {
	sr := NewScrapeRequest(url)

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(sr)

	res, _ := http.Post(api.endpoint, "application/json; charset=utf-8", b)
	fmt.Println(res.Body)
	body, _ := ioutil.ReadAll(res.Body)
	var respData ScrapeResponse
	json.Unmarshal(body, &respData)

	return &respData
}
