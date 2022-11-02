package nftstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	HTTPClient *http.Client
	APIKey     string
}

func (c Client) UploadFile(data []byte, mediaType string) (string, error) {
	req, _ := http.NewRequest("POST", "https://api.nft.storage/upload", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Accept", "application/json")

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var resBody responseBody
	if err := json.NewDecoder(response.Body).Decode(&resBody); err != nil {
		return "", err
	}

	if !resBody.Ok || response.StatusCode != 200 {
		return "", fmt.Errorf("non 'Ok' response from nftstorage. code=%d error=%v", response.StatusCode, resBody.Error)
	}

	return resBody.Value.Cid, nil
}

type responseBody struct {
	Ok    bool `json:"ok"`
	Value struct {
		Cid string `json:"cid"`
	} `json:"value"`
	Error struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
}
