package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type searchResult struct {
	Kind  string `json:"kind"`
	Items []struct {
		Image struct {
			ContextLink   string `json:"contextLink"`
			ThumbnailLink string `json:"thumbnailLink"`
		} `json:"image"`
	} `json:"items"`
}

func findImages(query string) (*searchResult, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/customsearch/v1", nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("searchType", "image")
	q.Add("cx", os.Getenv("GOOGLE_SEARCH_CX"))
	q.Add("key", os.Getenv("GOOGLE_SEARCH_KEY"))
	req.URL.RawQuery = q.Encode()

	resp, err := (&http.Client{}).Do(req)

	if err != nil {
		return nil, err
	}

	result := searchResult{}

	err = json.NewDecoder(resp.Body).Decode(&result)

	return &result, err
}
