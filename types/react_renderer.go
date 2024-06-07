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
	htmlTpl "html/template"
	"strings"
	textTpl "text/template"

	"github.com/mkloubert/go-package-manager/resources"
	"github.com/mkloubert/go-package-manager/utils"
)

// ReactRenderer helps to render HTML output
// with JSX and React
type ReactRenderer struct {
	BodyClass       string                                 // the className for the body tag
	ContentClass    string                                 // the className for the content wrapper
	ExternalModules map[string]ReactRendererExternalModule // list of external modules
	JsModules       [][]byte                               // list of contents of JavaScript modules to include
	Jsx             [][]byte                               // list of contents of JSX modules to include
	Template        string                                 // the name of the custom template inside resources to use
	Vars            map[string]interface{}                 // variables to inject into the final HTML as JavaScript variables
}

// ReactRendererExternalModule describes an external module
type ReactRendererExternalModule struct {
	Type string // the type, like "module"
	Url  string // the URL
}

// AddExternalEsmModule() - adds an external ESM module
func (rr *ReactRenderer) AddExternalEsmModule(name string, url string) {
	url = strings.TrimSpace(url)

	externalModules := map[string]ReactRendererExternalModule{}
	if rr.ExternalModules != nil {
		externalModules = rr.ExternalModules // use from renderer
	}

	externalModules[name] = ReactRendererExternalModule{
		Type: "module",
		Url:  strings.TrimSpace(url),
	}

	rr.ExternalModules = externalModules
}

// AddJavascriptTemplate() - adds a JavaScript template by name from resources
func (rr *ReactRenderer) AddJavascriptTemplate(name string) error {
	jsData, err := resources.JavaScripts.ReadFile(
		fmt.Sprintf("javascripts/%v.js", name),
	)

	if err != nil {
		return err
	}

	jsModules := [][]byte{}
	if rr.JsModules != nil {
		jsModules = rr.JsModules
	}

	jsModules = append(jsModules, jsData)
	rr.JsModules = jsModules

	return err
}

// AddJsxTemplate() - adds a JSX template by name from resources
func (rr *ReactRenderer) AddJsxTemplate(name string) error {
	jsxData, err := resources.JSX.ReadFile(
		fmt.Sprintf("jsx/%v.jsx", name),
	)

	if err != nil {
		return err
	}

	jsx := [][]byte{}
	if rr.Jsx != nil {
		jsx = rr.Jsx
	}

	jsx = append(jsx, jsxData)
	rr.Jsx = jsx

	return err
}

// AddVars() - add one or more variable for the output HTML
func (rr *ReactRenderer) AddVars(varsToAdd map[string]interface{}) {
	vars := map[string]interface{}{}
	if rr.Vars != nil {
		vars = rr.Vars
	}

	for k, v := range varsToAdd {
		vars[k] = v
	}

	rr.Vars = vars
}

// Render() - renders the final HTML using a name
func (rr *ReactRenderer) Render(name string) ([]byte, error) {
	template := strings.TrimSpace(rr.Template)
	if template == "" {
		template = "default"
	}

	// React.js
	reactJSCodeData, err := resources.JavaScripts.ReadFile("javascripts/react@18.3.1.min.js")
	if err != nil {
		return []byte{}, err
	}

	// ReactDOM
	reactDOMJSCodeData, err := resources.JavaScripts.ReadFile("javascripts/react-dom@18.3.1.min.js")
	if err != nil {
		return []byte{}, err
	}

	// Babel.js
	babelJSCodeData, err := resources.JavaScripts.ReadFile("javascripts/babel@7.24.6.min.js")
	if err != nil {
		return []byte{}, err
	}

	// global
	globalsJSCodeData, err := resources.JavaScripts.ReadFile("javascripts/globals.js")
	if err != nil {
		return []byte{}, err
	}

	// global React hooks
	hooksJSCodeData, err := resources.JavaScripts.ReadFile("javascripts/hooks.jsx")
	if err != nil {
		return []byte{}, err
	}

	// the React template to load
	templateData, err := resources.Templates.ReadFile(fmt.Sprintf("templates/react/%v.html", template))
	if err != nil {
		return []byte{}, err
	}

	// start with text template
	tpl, err := textTpl.New(name).Parse(string(templateData))
	if err != nil {
		return []byte{}, err
	}

	// Base64 encoded Data URIs of JavaScript modules to include
	jsModuleList := []interface{}{}
	if rr.JsModules != nil {
		for _, jsData := range rr.JsModules {
			jsModuleList = append(jsModuleList, map[string]interface{}{
				"CodeBase64": utils.ToDataUri(jsData, "text/javascript"),
			})
		}
	}

	// Base64 encoded Data URIs of JSX code to include
	jsxCodeList := []interface{}{}
	if rr.Jsx != nil {
		for _, jsxData := range rr.Jsx {
			jsxCodeList = append(jsxCodeList, map[string]interface{}{
				"CodeBase64": utils.ToDataUri(jsxData, "text/babel"),
			})
		}
	}

	// collect variables to include into final HTML
	vars := map[string]interface{}{}
	if rr.Vars != nil {
		for k, v := range rr.Vars {
			jsonStr := "null"
			if v != nil {
				jsonStrData, err := json.Marshal(v)
				utils.CheckForError(err)

				jsonStr = string(jsonStrData)
			}

			vars[k] = jsonStr
		}
	}

	data := map[string]interface{}{
		"BabelJSCodeBase64":    utils.ToDataUri(babelJSCodeData, "text/javascript"),
		"BodyClass":            strings.TrimSpace(rr.BodyClass),
		"ContentClass":         strings.TrimSpace(rr.ContentClass),
		"ExternalModules":      rr.ExternalModules,
		"GlobalsJSCodeBase64":  utils.ToDataUri(globalsJSCodeData, "text/javascript"),
		"HooksJSCodeBase64":    utils.ToDataUri(hooksJSCodeData, "text/babel"),
		"JSModuleList":         jsModuleList,
		"JSXCodeList":          jsxCodeList,
		"ReactDOMJSCodeBase64": utils.ToDataUri(reactDOMJSCodeData, "text/javascript"),
		"ReactJSCodeBase64":    utils.ToDataUri(reactJSCodeData, "text/javascript"),
		"VariablesJSONList":    vars,
	}

	htmlTpl.JSEscaper()

	var htmlBuffer bytes.Buffer

	err = tpl.ExecuteTemplate(&htmlBuffer, name, data)
	if err != nil {
		return []byte{}, err
	}

	return htmlBuffer.Bytes(), nil
}
