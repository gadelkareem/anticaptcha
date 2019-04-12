package anticaptcha

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL      = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
	sendInterval = 10 * time.Second
)

type Client struct {
	APIKey string
}

// Method to create the task to process the recaptcha, returns the task_id
func (c *Client) createTaskRecaptcha(websiteURL string, recaptchaKey string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":       "NoCaptchaTaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	if responseBody["taskId"] == nil {
		return 0, errors.Errorf("Failed to get a response")
	}
	return responseBody["taskId"].(float64), nil
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Client) getTaskResult(taskID float64) (map[string]interface{}, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, nil
}

// SendRecaptcha Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendRecaptcha(websiteURL string, recaptchaKey string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskRecaptcha(websiteURL, recaptchaKey)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			log.Println("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(sendInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			log.Println("Result is ready.")
			break
		}
	}

	if response["solution"] == nil {
		return "", errors.Errorf("Failed to get a response")
	}
	return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
}

// Method to create the task to process the image captcha, returns the task_id
func (c *Client) createTaskImage(imgString string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	return responseBody["taskId"].(float64), nil
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendImage(imgString string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskImage(imgString)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			log.Println("Result is not ready, waiting a few seconds to check again...")
			time.Sleep(sendInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			log.Println("Result is ready.")
			break
		}
	}
	return response["solution"].(map[string]interface{})["text"].(string), nil
}
