package routeprops

// Erp is what the Convexa middlewares put on the props of every request,
// beside what the backoffice ones put on Backoffice.
type Erp struct {
	// Language is the language the request asked for explicitly — the lang
	// query parameter, or the convexa_lang cookie a former one left —
	// normalized to "en" or "pt" by the erp-language middleware; "" when none.
	Language string
	// AcceptLanguage is the first language of the Accept-Language header
	// Convexa speaks, "en" or "pt"; "" when it lists none.
	AcceptLanguage string
	// Path is the path of the request, which the language switcher of a page
	// returns to.
	Path string
}
