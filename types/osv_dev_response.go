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

// OsvDevResponse stores information about a successful response
// from osv.dev API
type OsvDevResponse struct {
	Vulnerabilities *[]OsvDevResponseVulnerabilityItem `json:"vulns,omitempty"` // list of vulnerabilities
}

// OsvDevResponseVulnerabilityItem represents an item
// in OsvDevResponse.Vulnerabilities array
type OsvDevResponseVulnerabilityItem struct {
	DatabaseSpecific *OsvDevResponseVulnerabilityItemDataSpecificInfo `json:"database_specific,omitempty"` // database specific information
	Details          string                                           `json:"details,omitempty"`           // details
	Id               string                                           `json:"id,omitempty"`                // ID
	ModifiedDate     string                                           `json:"modified,omitempty"`          // modification date
	PublishedDate    string                                           `json:"published,omitempty"`         // publish date
	References       *[]OsvDevResponseVulnerabilityItemReference      `json:"references,omitempty"`        // list of references
	Severity         *[]OsvDevResponseVulnerabilitySeverityItem       `json:"severity,omitempty"`          // list of severities
	Summary          string                                           `json:"summary,omitempty"`           // summary
}

// OsvDevResponseVulnerabilityItemDataSpecificInfo represents value
// in OsvDevResponseVulnerabilityItem.DatabaseSpecific property
type OsvDevResponseVulnerabilityItemDataSpecificInfo struct {
	Severity string `json:"severity,omitempty"` // the severity
}

// OsvDevResponseVulnerabilityItemReference represents an item
// in OsvDevResponseVulnerabilityItem.References array
type OsvDevResponseVulnerabilityItemReference struct {
	Type string `json:"type,omitempty"` // the type
	Url  string `json:"url,omitempty"`  // the URL
}

// OsvDevResponseVulnerabilitySeverityItem represents an item
// in OsvDevResponseVulnerabilityItem.v array
type OsvDevResponseVulnerabilitySeverityItem struct {
	Score string `json:"score,omitempty"` // the score
	Type  string `json:"type,omitempty"`  // the type
}
