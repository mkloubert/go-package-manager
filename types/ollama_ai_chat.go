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

	"github.com/mkloubert/go-package-manager/utils"
)

// OllamaAIChat is an implementation of ChatAI interface
// using local Ollama REST API
type OllamaAIChat struct {
	Conversation []OllamaAIChatMessage // the conversation
	Model        string                // the current model
	SystemPrompt string                // the current system prompt
	Temperature  float32               // the current temperature
	Verbose      bool                  // running in verbose mode or not
}

// OllamaAIChatMessage is an item inside
// OllamaAIChat.Conversation array
type OllamaAIChatMessage struct {
	Content string `json:"content,omitempty"` // the message content
	Role    string `json:"role,omitempty"`    // the role like user, assistant or system
}

// OllamaApiResponse is the data of a successful chat conversation response
type OllamaApiChatCompletionResponse struct {
	Message OllamaAIChatMessage `json:"message,omitempty"` // the message
}

func (c *OllamaAIChat) AddToHistory(role string, content string) {
	c.Conversation = append(c.Conversation, OllamaAIChatMessage{
		Content: content,
		Role:    role,
	})
}

func (c *OllamaAIChat) ClearHistory() {
	c.Conversation = []OllamaAIChatMessage{}
}

func (c *OllamaAIChat) DescribeImage(message string, dataURI string) (DescribeImageResponse, error) {
	var imageDescription DescribeImageResponse

	base64Content, err := utils.Base64FromDataURI(dataURI)
	if err != nil {
		return imageDescription, err
	}

	url := "http://localhost:11434/api/chat"

	messages := []map[string]interface{}{}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": message,
		"images":  []string{base64Content},
	})

	body := map[string]interface{}{
		"model":  c.GetModel(),
		"stream": false,
		"format": map[string]interface{}{
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
		"messages": messages,
	}

	if c.SystemPrompt != "" {
		body["system"] = c.SystemPrompt
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
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return imageDescription, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return imageDescription, fmt.Errorf("unexpected response: %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return imageDescription, err
	}

	var completionResponse OllamaApiChatCompletionResponse
	err = json.Unmarshal(responseData, &completionResponse)
	if err != nil {
		return imageDescription, err
	}

	return get_ai_image_description_from_json(completionResponse.Message.Content)
}

func (c *OllamaAIChat) GetModel() string {
	return c.Model
}

func (c *OllamaAIChat) GetMoreInfo() string {
	return ""
}

func (c *OllamaAIChat) GetPromptSuffix() string {
	return ""
}

func (c *OllamaAIChat) GetProvider() string {
	return "ollama"
}

func (c *OllamaAIChat) SendMessage(message string, onUpdate ChatAIMessageChunkReceiver) error {
	url := "http://localhost:11434/api/chat"

	userMessage := OllamaAIChatMessage{
		Content: message,
		Role:    "user",
	}

	messages := []OllamaAIChatMessage{}
	messages = append(messages, c.Conversation...)
	messages = append(messages, userMessage)

	body := map[string]interface{}{
		"model":       c.Model,
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

	var chatResponse OllamaApiChatCompletionResponse
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return err
	}

	assistantMessage := OllamaAIChatMessage{
		Content: chatResponse.Message.Content,
		Role:    chatResponse.Message.Role,
	}

	c.Conversation = append(
		c.Conversation,
		userMessage, assistantMessage,
	)

	return onUpdate(assistantMessage.Content)
}

func (c *OllamaAIChat) SendPrompt(prompt string, onUpdate ChatAIMessageChunkReceiver) error {
	var systemMessage *string
	if c.SystemPrompt != "" {
		systemMessage = &c.SystemPrompt
	}

	url := "http://localhost:11434/api/generate"

	body := map[string]interface{}{
		"model":       c.Model,
		"prompt":      prompt,
		"stream":      false,
		"temperature": c.Temperature,
	}
	if systemMessage != nil {
		body["system"] = *systemMessage
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
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response: %v", resp.StatusCode)
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var completionResponse OllamaApiCompletionResponse
	err = json.Unmarshal(responseData, &completionResponse)
	if err != nil {
		return err
	}

	onUpdate(completionResponse.Response)
	return nil
}

func (c *OllamaAIChat) UpdateModel(modelName string) {
	c.Model = strings.TrimSpace(modelName)
}

func (c *OllamaAIChat) UpdateSystem(systemPrompt string) {
	c.SystemPrompt = systemPrompt

	c.Conversation = []OllamaAIChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}
}

func (c *OllamaAIChat) UpdateTemperature(newValue float32) {
	c.Temperature = newValue
}

func (c *OllamaAIChat) WithJsonSchema(message string, schemaName string, schema map[string]interface{}, onUpdate ChatAIMessageChunkReceiver) error {
	model := strings.TrimSpace(strings.ToLower(c.Model))
	if model == "" {
		return fmt.Errorf("no chat ai model defined")
	}

	url := "http://localhost:11434/api/chat"

	userMessage := OllamaAIChatMessage{
		Content: message,
		Role:    "user",
	}

	messages := []OllamaAIChatMessage{}

	if c.SystemPrompt != "" {
		systemMessage := OllamaAIChatMessage{
			Content: c.SystemPrompt,
			Role:    "system",
		}

		messages = append(messages, systemMessage)
	}

	messages = append(messages, c.Conversation...)
	messages = append(messages, userMessage)

	body := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"stream":      false,
		"temperature": c.Temperature,
		"format":      schema,
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

	var chatResponse OllamaApiChatCompletionResponse
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return err
	}

	assistantMessage := OllamaAIChatMessage{
		Content: chatResponse.Message.Content,
		Role:    chatResponse.Message.Role,
	}

	c.Conversation = append(
		c.Conversation,
		userMessage, assistantMessage,
	)

	return onUpdate(assistantMessage.Content)
}
