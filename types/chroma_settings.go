package types

import (
	"github.com/alecthomas/chroma/quick"
)

// ChromaSettings stores settings for syntax highlighted console output
type ChromaSettings struct {
	app       *AppContext
	Formatter string // name of the formatter
	Style     string // name of the style
}

// app.Highlight() - tries to output a string highlighted in the defined language
func (cs *ChromaSettings) Highlight(s string, language string) {
	err := quick.Highlight(cs.app, s, language, cs.Formatter, cs.Style)
	if err != nil {
		cs.app.Write([]byte(s))
	}
}
