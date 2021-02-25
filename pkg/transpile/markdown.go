package transpile

import (
	"io"
	"io/ioutil"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// Markdown formats the provided bytes using whatever Markdown engine we're currently using.
func Markdown(input io.Reader, output io.Writer) error {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.DefinitionList,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	buf, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	md.Convert(buf, output)
	return nil
}
