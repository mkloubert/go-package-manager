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

// app.Write() - implementation for an io.Writer
func (app *AppContext) Write(p []byte) (int, error) {
	out := app.Out
	if out == nil {
		return len(p), nil // deactivated
	}

	return out.Write(p)
}

// app.WriteError() - writes to error output
func (app *AppContext) WriteError(p []byte) (int, error) {
	errorOut := app.ErrorOut
	if errorOut == nil {
		return len(p), nil // deactivated
	}

	return errorOut.Write(p)
}

// app.WriteErrorString() - writes to error output
func (app *AppContext) WriteErrorString(s string) (int, error) {
	return app.WriteError([]byte(s))
}

// app.WriteString() - implementation for an io.StringWriter
func (app *AppContext) WriteString(s string) (int, error) {
	return app.Write([]byte(s))
}
