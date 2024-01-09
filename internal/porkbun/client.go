// Copyright (c) Saba Gogichaishvili
// SPDX-License-Identifier: ISC

package porkbun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	Credentials
	url    *url.URL
	client *http.Client
}

func NewClient(
	client *http.Client,
	apiKey string,
	secretAPIKey string,
	forceIPv4 bool,
) *Client {
	host := "porkbun.com"
	if forceIPv4 {
		host = "api-ipv4.porkbun.com"
	}
	scheme := "https"
	path := "api/json/v3"
	return &Client{
		Credentials: Credentials{
			APIKey:       apiKey,
			SecretAPIKey: secretAPIKey,
		},
		url: &url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   path,
		},
		client: client,
	}
}

func (c *Client) get(
	ctx context.Context,
	path string,
	res status,
) error {
	return c.do(ctx, http.MethodGet, path, nil, res)
}

func (c *Client) post(
	ctx context.Context,
	path string,
	req credentials,
	res status,
) error {
	req.setAPIKey(c.APIKey)
	req.setSecretAPIKey(c.SecretAPIKey)
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf(
			"Failed to marshal an API request as JSON "+
				"with the following error: %s",
			err,
		)
	}

	return c.do(ctx, http.MethodPost, path, bytes.NewReader(body), res)
}

func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	response status,
) error {
	url := c.url.JoinPath(path).String()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf(
			"Failed to create an HTPP request "+
				"with the following error: %s",
			err,
		)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"Failed to process an HTTP request "+
				"with the following error: %s",
			err,
		)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf(
			"Failed to read response body "+
				"with the following error: %s",
			err,
		)
	}

	err = json.Unmarshal(b, response)
	if err != nil {
		return fmt.Errorf(
			"Failed to unmarshal an API response as JSON "+
				"with the following error: %s",
			err,
		)
	}

	switch response.getStatus() {
	case "SUCCESS":
		return nil
	case "ERROR":
		return fmt.Errorf(
			"Received 'ERROR' response status "+
				"with the following message: %s",
			response.getMessage(),
		)
	default:
		return fmt.Errorf(
			"Received unknown response status. "+
				"Expected 'SUCCESS' or 'ERROR', got: '%s'.",
			response.getStatus(),
		)
	}
}
