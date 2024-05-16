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

// OpenAIChatCompletionResponseV1 stores data of a successful
// OpenAI chat completion response API response (version 1)
type OpenAIChatCompletionResponseV1 struct {
	Choices []OpenAIChatCompletionResponseV1Choice `json:"choices"` // list of choices
	Model   string                                 `json:"model"`   // used model
	Usage   OpenAIChatCompletionResponseV1Usage    `json:"usage"`   // the usage
}

// OpenAIChatCompletionResponseV1Choice is an item inside `choices` property
// of an `OpenAIChatCompletionResponseV1` object
type OpenAIChatCompletionResponseV1Choice struct {
	Index   int32                                       `json:"index"`   // the zero-based index
	Message OpenAIChatCompletionResponseV1ChoiceMessage `json:"message"` // the message information
}

// OpenAIChatCompletionResponseV1ChoiceMessage contains data for `message` property
// of an `OpenAIChatCompletionResponseV1ChoiceMessage` object
type OpenAIChatCompletionResponseV1ChoiceMessage struct {
	Content string `json:"content"` // the message context
	Role    string `json:"role"`    // the role like 'user' or 'assistant'
}

// OpenAIChatCompletionResponseV1Usage contains data for `usage` property
// of an `OpenAIChatCompletionResponseV1` object
type OpenAIChatCompletionResponseV1Usage struct {
	CompletionTokens int32 `json:"completion_tokens"` // number of completion tokens
	PromptTokens     int32 `json:"prompt_tokens"`     // number of prompt tokens
	TotalTokens      int32 `json:"total_tokens"`      // number of total used tokens
}
