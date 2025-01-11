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
	"encoding/json"
	"strings"
)

// ChatAI describes an object that provides abstract
// methods to interaction with a chat API
type ChatAI interface {
	// ChatAI.ClearHistory() - clears chat history
	ClearHistory()
	// ChatAI.DescribeImage() - describes an image without adding using history
	DescribeImage(message string, dataURI string) (DescribeImageResponse, error)
	// ChatAI.GetModel() - get the name of the chat model
	GetModel() string
	// ChatAI.GetMoreInfo() - returns additional information, if available
	GetMoreInfo() string
	// ChatAI.GetPromptSuffix() - returns additional suffix for an input prompt, if available
	GetPromptSuffix() string
	// ChatAI.GetProvider() - get the name of the chat provider
	GetProvider() string
	// ChatAI.SendMessage() - sends a new message
	// to the API for the current chat conversation
	SendMessage(message string, onUpdate ChatAIMessageChunkReceiver) error
	// ChatAI.SendPrompt() - sends a single completion prompt
	SendPrompt(prompt string, onUpdate ChatAIMessageChunkReceiver) error
	// ChatAI.SendMessage() - switches the model
	UpdateModel(modelName string)
	// ChatAI.UpdateSystem() - clears chat history and sets the
	// system prompt
	UpdateSystem(systemPromt string)
	// ChatAI.UpdateSystem() - sets up new temperature value
	UpdateTemperature(newValue float32)
	// WithJsonSchema() - sends a message with a JSON schema
	WithJsonSchema(message string, schemaName string, schema map[string]interface{}, onUpdate ChatAIMessageChunkReceiver) error
}

type ChatAIMessageChunkReceiver = func(messageChunk string) error

func get_ai_image_description_from_json(jsonStr string) (DescribeImageResponse, error) {
	var imageDescription DescribeImageResponse

	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(strings.TrimSpace(jsonStr)), &data)
	if err != nil {
		return imageDescription, err
	}

	aria_attributes, ok := data["aria_attributes"]
	if ok {
		attributes, ok := aria_attributes.(map[string]interface{})
		if ok {
			aria_description, ok := attributes["aria_description"]
			if ok {
				description, ok := aria_description.(string)
				if ok {
					imageDescription.Description = strings.TrimSpace(description)
				}
			}

			aria_label, ok := attributes["aria_label"]
			if ok {
				label, ok := aria_label.(string)
				if ok {
					imageDescription.Label = strings.TrimSpace(label)
				}
			}
		}
	}

	return imageDescription, nil
}
