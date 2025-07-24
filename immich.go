package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (u *User) immichRequest(ctx context.Context, method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.ImmichURL+"api/"+url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", u.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad response status code %+v", *resp)
	}

	data, err := io.ReadAll(resp.Body)

	return data, err
}
