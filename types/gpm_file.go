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
	"os"

	"github.com/goccy/go-yaml"
)

// A GpmFile stores all data of a gpm.y(a)ml file.
type GpmFile struct {
	Files   []string          `yaml:"files,omitempty"`   // whitelist of file patterns which are used by pack command for example
	Name    string            `yaml:"name,omitempty"`    // the name
	Scripts map[string]string `yaml:"scripts,omitempty"` // one or more scripts
}

// LoadGpmFile() - Loads a gpm.yaml file via a file path
func LoadGpmFile(gpmFilePath string) (GpmFile, error) {
	var gpm GpmFile
	defer func() {
		if gpm.Files == nil {
			gpm.Files = []string{}
		}
		if gpm.Scripts == nil {
			gpm.Scripts = map[string]string{}
		}
	}()

	yamlData, err := os.ReadFile(gpmFilePath)
	if err != nil {
		return gpm, err
	}

	err = yaml.Unmarshal(yamlData, &gpm)

	return gpm, err
}
