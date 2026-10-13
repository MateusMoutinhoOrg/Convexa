package gohtmlpage

// The standard adapter's html page implementation: the factory filling
// deps.Deps.PageDeps out of the assets compiled into the binary. Like goembed,
// it reaches for the assets package outside the sandbox; the sandbox only ever
// sees the function field of pagedeps.Contract.

import (
	"bytes"
	htmltemplate "html/template"
	"path"

	"github.com/MateusMoutinhoOrg/Convexa/assets"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/deps"
	pagedeps "github.com/MateusMoutinhoOrg/Convexa/sandbox/deps/pagedeps"
)

// renderHTMLPage fills pagedeps.Contract.RenderHTMLPage: every template of
// paths parsed into one set, in order, and the one named entry executed with
// vars.
func renderHTMLPage(paths []string, entry string, vars interface{}) ([]byte, error) {
	set := htmltemplate.New(entry)
	for _, requested := range paths {
		content, err := assets.Files.ReadFile(path.Join(".", requested))
		if err != nil {
			return nil, err
		}
		_, err = set.New(path.Base(requested)).Parse(string(content))
		if err != nil {
			return nil, err
		}
	}
	var buf bytes.Buffer
	if err := set.ExecuteTemplate(&buf, entry, vars); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Bind fills deps.Deps.PageDeps with html/template over the project's assets.
func Bind(deps *deps.Deps) {
	deps.PageDeps = pagedeps.Contract{
		RenderHTMLPage: func(paths []string, entry string, vars interface{}) ([]byte, error) {
			return renderHTMLPage(paths, entry, vars)
		},
	}
}
