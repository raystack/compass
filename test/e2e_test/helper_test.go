//go:build e2e
// +build e2e

package endtoend_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/raystack/compass/core/asset"
)

var (
	SERVER_HOST               = "http://localhost:8080"
	IDENTITY_HEADER_KEY_UUID  = "Compass-User-UUID"
	IDENTITY_HEADER_KEY_EMAIL = "Compass-User-Email"
)

// Client is the http client implementation
type Client struct {
	client  *http.Client
	host    string
	headers map[string]string
}

func NewClient() *Client {
	client := Client{
		client: &http.Client{},
		host:   SERVER_HOST,
	}

	return &client
}

func (c *Client) PatchAsset(ast asset.Asset) (string, error) {
	path := "/v1beta1/assets"
	url := c.host + path
	type requestStruct struct {
		Asset asset.Asset `json:"asset"`
	}
	type responseStruct struct {
		ID string
	}

	var responsePayload responseStruct
	requestPayload := requestStruct{
		Asset: ast,
	}
	err := c.makeRequest(http.MethodPatch, url, &requestPayload, &responsePayload)
	if err != nil {
		return "", err
	}

	return responsePayload.ID, nil
}

func (c *Client) GetAllAssets() ([]asset.Asset, error) {
	path := "/v1beta1/assets?size=5"
	url := c.host + path
	type responseStruct struct {
		Data []asset.Asset
	}

	var responsePayload responseStruct
	err := c.makeRequest(http.MethodGet, url, nil, &responsePayload)
	if err != nil {
		return nil, err
	}
	return responsePayload.Data, nil
}

func (c *Client) GetAnAsset(id string) (asset.Asset, error) {
	path := "/v1beta1/assets/" + id
	url := c.host + path
	type responseStruct struct {
		Data asset.Asset `json:"data"`
	}

	var responsePayload responseStruct
	err := c.makeRequest(http.MethodGet, url, nil, &responsePayload)
	if err != nil {
		return asset.Asset{}, err
	}
	return responsePayload.Data, nil
}

func (c *Client) GetAssetVersions(id string) ([]asset.Asset, error) {
	path := "/v1beta1/assets/" + id + "/versions"
	url := c.host + path
	type responseStruct struct {
		Data []asset.Asset `json:"data"`
	}

	var responsePayload responseStruct
	err := c.makeRequest(http.MethodGet, url, nil, &responsePayload)
	if err != nil {
		return nil, err
	}
	return responsePayload.Data, nil
}

func (c *Client) GetAssetWithVersion(id, version string) (asset.Asset, error) {
	path := "/v1beta1/assets/" + id + "/versions/" + version
	url := c.host + path
	type responseStruct struct {
		Data asset.Asset `json:"data"`
	}

	var responsePayload responseStruct
	err := c.makeRequest(http.MethodGet, url, nil, &responsePayload)
	if err != nil {
		return asset.Asset{}, err
	}
	return responsePayload.Data, nil
}

func (c *Client) makeRequest(method, url string, payload interface{}, data interface{}) (err error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode the payload JSON: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set(IDENTITY_HEADER_KEY_UUID, "compassendtoendtest@raystack.io")
	req.Header.Set(IDENTITY_HEADER_KEY_EMAIL, "compassendtoendtest@raystack.io")

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to generate response")
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body")
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("getting %d status code, body: %s", res.StatusCode, string(bytes))
	}

	if err = json.Unmarshal(bytes, &data); err != nil {
		return fmt.Errorf("failed to parse: %s, err: %w", string(bytes), err)
	}
	return
}
