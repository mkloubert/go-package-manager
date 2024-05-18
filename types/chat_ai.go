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

// ChatAI describes an object that provides abstract
// methods to interaction with a chat API
type ChatAI interface {
	// ChatAI.ClearHistory() - clears chat history
	ClearHistory()
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
	// ChatAI.SendMessage() - switches the model
	UpdateModel(modelName string)
	// ChatAI.UpdateSystem() - clears chat history and sets the
	// system prompt
	UpdateSystem(systemPromt string)
	// ChatAI.UpdateSystem() - sets up new temperature value
	UpdateTemperature(newValue float32)
}

type ChatAIMessageChunkReceiver = func(messageChunk string) error
