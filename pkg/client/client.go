package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		logger: log.New(log.Writer(), "GoKV Client: ", log.LstdFlags),
	}
}

func (c *Client) Get(key string) (interface{}, error) {
	c.logger.Printf("Getting value for key: %s", key)
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/get/%s", c.baseURL, key))
	if err != nil {
		c.logger.Printf("Error getting value for key %s: %v", key, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.logger.Printf("Key not found: %s", key)
		return nil, fmt.Errorf("key not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Printf("Error decoding response for key %s: %v", key, err)
		return nil, err
	}

	c.logger.Printf("Successfully retrieved value for key: %s", key)
	return result["value"], nil
}

func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	c.logger.Printf("Setting value for key: %s", key)
	data := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	if ttl > 0 {
		data["ttl"] = ttl.Seconds()
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Printf("Error marshaling data for key %s: %v", key, err)
		return err
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s/set", c.baseURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Printf("Error setting value for key %s: %v", key, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	c.logger.Printf("Successfully set value for key: %s", key)
	return nil
}

func (c *Client) Delete(key string) error {
	c.logger.Printf("Deleting key: %s", key)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/delete/%s", c.baseURL, key), nil)
	if err != nil {
		c.logger.Printf("Error creating delete request for key %s: %v", key, err)
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Printf("Error deleting key %s: %v", key, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	c.logger.Printf("Successfully deleted key: %s", key)
	return nil
}

func (c *Client) Query(queryString string) (interface{}, error) {
	c.logger.Printf("Executing query: %s", queryString)
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/query?q=%s", c.baseURL, queryString))
	if err != nil {
		c.logger.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Printf("Error decoding query response: %v", err)
		return nil, err
	}

	c.logger.Printf("Successfully executed query")
	return result, nil
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	var errorResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
		if errorMsg, ok := errorResp["error"]; ok {
			c.logger.Printf("Server error: %s", errorMsg)
			return fmt.Errorf("server error: %s", errorMsg)
		}
	}
	c.logger.Printf("Unexpected status code: %d", resp.StatusCode)
	return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
