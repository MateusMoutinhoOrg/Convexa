package pagedeps

// This package is the sandbox's *copy* of the api an html page library
// exposes — the same mechanic as embeddeps, for the same reason: parsing a
// template is html/template's work, and the assets it reads are compiled into
// the binary by the `embed` directive, so neither may appear inside the
// sandbox. The contract is restated here, and the adapter — which lives
// outside the sandbox — is what fills it.
//
// Where embeddeps.RenderHTMLTemplate parses one file, a page here is several
// files parsed into one template set: a layout holding what every page shares
// — the head, the styles, the navigation — and the page itself, which defines
// the blocks the layout leaves open. Every path is slash-separated and
// relative to the root of the asset tree, as in embeddeps.

// Contract is the html page library injected whole as the Deps.PageDeps field.
type Contract struct {
	// RenderHTMLPage parses every template of paths, in order, into one
	// set, executes the template named entry with vars and returns the
	// result. Every value is escaped for the context it lands in — text, an
	// attribute, a url — so a value nobody escaped by hand can never become
	// markup. The error reports a template that is missing, does not parse,
	// or fails while executing.
	RenderHTMLPage func(paths []string, entry string, vars interface{}) ([]byte, error)
}
