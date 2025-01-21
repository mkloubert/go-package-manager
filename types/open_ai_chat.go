// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIChat is an implementation of ChatAI interface
// using remote ChatGPT REST API by OpenAI
type OpenAIChat struct {
	ApiKey       string              // the API key to use
	Conversation []OpenAIChatMessage // the conversation
	Model        string              // the current model
	SystemPrompt string              // the current system prompt
	Temperature  float32             // the current temperature
	TotalTokens  int32               // number of total used tokens in this session
	Verbose      bool                // running in verbose mode or not
}

// OpenAIChatMessage is an item inside
// OpenAIChat.Conversation array
type OpenAIChatMessage struct {
	Content string `json:"content,omitempty"` // the message content
	Role    string `json:"role,omitempty"`    // the role like user, assistant or system
}

func (c *OpenAIChat) AddToHistory(role string, content string) {
	c.Conversation = append(c.Conversation, OpenAIChatMessage{
		Content: content,
		Role:    role,
	})
}

func (c *OpenAIChat) ClearHistory() {
	c.Conversation = []OpenAIChatMessage{}
}

func (c *OpenAIChat) DescribeImage(message string, dataURI string) (DescribeImageResponse, error) {
	var imageDescription DescribeImageResponse

	apiKey := strings.TrimSpace(c.ApiKey)
	if apiKey == "" {
		return imageDescription, fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.Model))
	if model == "" {
		return imageDescription, fmt.Errorf("no chat ai model defined")
	}

	url := "https://api.openai.com/v1/chat/completions"

	messages := []map[string]interface{}{}

	// system prompt
	if c.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": c.SystemPrompt,
		})
	}

	userContents := make([]map[string]interface{}, 0)
	userContents = append(userContents, map[string]interface{}{
		"type": "text",
		"text": message,
	})
	userContents = append(userContents, map[string]interface{}{
		"type": "image_url",
		"image_url": map[string]interface{}{
			"url": dataURI,
		},
	})

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContents,
	})

	body := map[string]interface{}{
		"model":    c.GetModel(),
		"messages": messages,
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name": "JSONAriaSchema",
				"schema": map[string]interface{}{
					"type":     "object",
					"required": []string{"aria_attributes"},
					"properties": map[string]interface{}{
						"aria_attributes": map[string]interface{}{
							"description": "HTML accessibility attributes which describe the image.",
							"type":        "object",
							"required":    []string{"aria_description", "aria_label"},
							"properties": map[string]interface{}{
								"aria_description": map[string]interface{}{
									"description": "Defines a string value that describes or annotates the image in detail.",
									"type":        "string",
								},
								"aria_label": map[string]interface{}{
									"description": "Defines a string value that can be used to name the image.",
									"type":        "string",
								},
							},
						},
					},
				},
			},
		},
		"stream":      false,
		"temperature": c.Temperature,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return imageDescription, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return imageDescription, err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return imageDescription, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return imageDescription, fmt.Errorf("unexpected response %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return imageDescription, err
	}

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return imageDescription, err
	}

	assistantMessage := OpenAIChatMessage{
		Content: "",
		Role:    "assistant",
	}
	if len(chatResponse.Choices) > 0 {
		assistantMessage.Content = chatResponse.Choices[0].Message.Content
		assistantMessage.Role = chatResponse.Choices[0].Message.Role
	}

	c.TotalTokens += chatResponse.Usage.TotalTokens

	return get_ai_image_description_from_json(assistantMessage.Content)
}

func (c *OpenAIChat) GetModel() string {
	return c.Model
}

func (c *OpenAIChat) GetMoreInfo() string {
	return fmt.Sprintf(
		"%vTotal tokens: %v",
		fmt.Sprintln(),
		c.TotalTokens,
	)
}

func (c *OpenAIChat) GetPromptSuffix() string {
	if c.Verbose {
		return fmt.Sprintf(" (%v)", c.TotalTokens)
	}

	return ""
}

func (c *OpenAIChat) GetProvider() string {
	return "openai"
}

func (c *OpenAIChat) SendMessage(message string, onUpdate ChatAIMessageChunkReceiver) error {
	apiKey := strings.TrimSpace(c.ApiKey)
	if apiKey == "" {
		return fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.Model))
	if model == "" {
		return fmt.Errorf("no chat ai model defined")
	}

	url := "https://api.openai.com/v1/chat/completions"

	userMessage := OpenAIChatMessage{
		Content: message,
		Role:    "user",
	}

	messages := []OpenAIChatMessage{}
	messages = append(messages, c.Conversation...)
	messages = append(messages, userMessage)

	body := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"stream":      false,
		"temperature": c.Temperature,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return err
	}

	assistantMessage := OpenAIChatMessage{
		Content: "",
		Role:    "assistant",
	}
	if len(chatResponse.Choices) > 0 {
		assistantMessage.Content = chatResponse.Choices[0].Message.Content
		assistantMessage.Role = chatResponse.Choices[0].Message.Role
	}

	c.Conversation = append(
		c.Conversation,
		userMessage, assistantMessage,
	)

	err = onUpdate(assistantMessage.Content)
	if err != nil {
		return err
	}

	c.TotalTokens += chatResponse.Usage.TotalTokens

	return nil
}

func (c *OpenAIChat) SendPrompt(prompt string, onUpdate ChatAIMessageChunkReceiver) error {
	apiKey := strings.TrimSpace(c.ApiKey)
	if apiKey == "" {
		return fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.Model))
	if model == "" {
		return fmt.Errorf("no chat ai model defined")
	}

	var systemMessage *OpenAIChatMessage
	if c.SystemPrompt != "" {
		systemMessage = &OpenAIChatMessage{
			Role:    "system",
			Content: c.SystemPrompt,
		}
	}

	userMessage := OpenAIChatMessage{
		Content: prompt,
		Role:    "user",
	}

	messages := []OpenAIChatMessage{}
	if systemMessage != nil {
		messages = append(messages, *systemMessage)
	}
	messages = append(messages, userMessage)

	url := "https://api.openai.com/v1/chat/completions"

	body := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"stream":      false,
		"temperature": c.Temperature,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return err
	}

	answer := ""
	if len(chatResponse.Choices) > 0 {
		answer = chatResponse.Choices[0].Message.Content
	}

	onUpdate(answer)
	return nil
}

func (c *OpenAIChat) UpdateModel(modelName string) {
	c.Model = strings.TrimSpace(modelName)
}

func (c *OpenAIChat) UpdateSystem(systemPrompt string) {
	c.SystemPrompt = systemPrompt

	c.Conversation = []OpenAIChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}
}

func (c *OpenAIChat) UpdateTemperature(newValue float32) {
	c.Temperature = newValue
}

func (c *OpenAIChat) WithJsonSchema(message string, schemaName string, schema map[string]interface{}, onUpdate ChatAIMessageChunkReceiver) error {
	apiKey := strings.TrimSpace(c.ApiKey)
	if apiKey == "" {
		return fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.Model))
	if model == "" {
		return fmt.Errorf("no chat ai model defined")
	}

	messages := []OpenAIChatMessage{}

	if c.SystemPrompt != "" {
		systemMessage := OpenAIChatMessage{
			Content: c.SystemPrompt,
			Role:    "system",
		}

		messages = append(messages, systemMessage)
	}

	userMessage := OpenAIChatMessage{
		Content: message,
		Role:    "user",
	}

	messages = append(messages, c.Conversation...)
	messages = append(messages, userMessage)

	url := "https://api.openai.com/v1/chat/completions"

	body := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   schemaName,
				"schema": schema,
			},
		},
		"temperature": c.Temperature,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return err
	}

	assistantMessage := OpenAIChatMessage{
		Content: "",
		Role:    "assistant",
	}
	if len(chatResponse.Choices) > 0 {
		assistantMessage.Content = chatResponse.Choices[0].Message.Content
		assistantMessage.Role = chatResponse.Choices[0].Message.Role
	}

	c.Conversation = append(
		c.Conversation,
		userMessage, assistantMessage,
	)

	err = onUpdate(assistantMessage.Content)
	if err != nil {
		return err
	}

	c.TotalTokens += chatResponse.Usage.TotalTokens

	return nil
}
