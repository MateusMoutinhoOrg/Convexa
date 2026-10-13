package erpi18n

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
)

// The languages Convexa speaks. Everything — the code, the api, the pages — is
// written in English, and English is the key every translation is looked up
// by: a text with no Portuguese entry is shown in English rather than
// missing. JSON keys and enum values of the api never change with the
// language; only the sentences meant for a person do.
const (
	// English is the language everything is written in, and the default.
	English = "en"
	// Portuguese is the language a user, a browser or an api call may ask for
	// instead.
	Portuguese = "pt"
)

// Languages is every language offered, in the order the switcher lists them.
func Languages(sandbox *api.Sandbox) []string {
	return []string{English, Portuguese}
}

// Normalize is the language value names — "pt", "pt-BR", "PT_br", "en-US" —
// or "" when it names none Convexa speaks.
func Normalize(sandbox *api.Sandbox, value string) string {
	value = sandbox.Deps.StringsDeps.ToLower(sandbox.Deps.StringsDeps.TrimSpace(value))
	if value == "pt" || sandbox.Deps.StringsDeps.HasPrefix(value, "pt-") || sandbox.Deps.StringsDeps.HasPrefix(value, "pt_") {
		return Portuguese
	}
	if value == "en" || sandbox.Deps.StringsDeps.HasPrefix(value, "en-") || sandbox.Deps.StringsDeps.HasPrefix(value, "en_") {
		return English
	}
	return ""
}

// FromAcceptLanguage is the first language of an Accept-Language header
// ("pt-BR,pt;q=0.9,en;q=0.8") Convexa speaks, or "" when it lists none. The
// header lists the languages most preferred first, so the weights are not
// read.
func FromAcceptLanguage(sandbox *api.Sandbox, header string) string {
	for _, part := range sandbox.Deps.StringsDeps.Split(header, ",") {
		tag := sandbox.Deps.StringsDeps.Split(part, ";")[0]
		if lang := Normalize(sandbox, tag); lang != "" {
			return lang
		}
	}
	return ""
}

// HtmlLang is the value of the lang attribute of a page in lang.
func HtmlLang(sandbox *api.Sandbox, lang string) string {
	if lang == Portuguese {
		return "pt-BR"
	}
	return "en"
}

// Name is how the language switcher names lang, in that language itself.
func Name(sandbox *api.Sandbox, lang string) string {
	if lang == Portuguese {
		return "Português"
	}
	return "English"
}

// Tr is text in lang: its Portuguese entry for Portuguese, text itself for
// English or for a text no entry translates.
func Tr(sandbox *api.Sandbox, lang string, text string) string {
	if lang != Portuguese {
		if spelled, ok := english[text]; ok {
			return spelled
		}
		return text
	}
	if translated, ok := portuguese[text]; ok {
		return translated
	}
	if translated, ok := fromFormat(sandbox, text); ok {
		return translated
	}
	return text
}

// fromFormat translates text that was formatted already — "The password
// needs at least 12 characters." — by the entry whose format it fits: every
// literal part of the format in order, a %d or a %s standing for the text
// between them, which is carried into the Portuguese format as it was.
func fromFormat(sandbox *api.Sandbox, text string) (string, bool) {
	for format, translated := range portuguese {
		if !sandbox.Deps.StringsDeps.Contains(format, "%") {
			continue
		}
		if args, ok := fit(sandbox, format, text); ok {
			return sandbox.Deps.StdDeps.Sprintf(sandbox.Deps.StringsDeps.ReplaceAll(translated, "%d", "%s"), args...), true
		}
	}
	return "", false
}

// fit is what stands for each %d and %s of format in text, or false when
// text does not fit format.
func fit(sandbox *api.Sandbox, format string, text string) ([]any, bool) {
	args := []any{}
	for len(format) > 0 {
		verb := indexOf(format, "%")
		if verb < 0 {
			return args, format == text
		}
		if verb+1 >= len(format) || len(text) < verb || text[:verb] != format[:verb] {
			return nil, false
		}
		text, format = text[verb:], format[verb+2:]
		next := indexOf(format, "%")
		literal := format
		if next >= 0 {
			literal = format[:next]
		}
		if literal == "" {
			if next >= 0 {
				return nil, false
			}
			args = append(args, text)
			return args, len(text) > 0
		}
		end := indexOf(text, literal)
		if end <= 0 {
			return nil, false
		}
		args = append(args, text[:end])
		text = text[end:]
	}
	return args, text == ""
}

// indexOf is where sub first starts in text, -1 when it does not.
func indexOf(text string, sub string) int {
	for i := 0; i+len(sub) <= len(text); i++ {
		if text[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// english spells, for a page in English, the enum values whose key reads
// worse than words: the api keeps the key, a page shows the words.
var english = map[string]string{
	"transfer-in":  "transfer in",
	"transfer-out": "transfer out",
	"credit-card":  "credit card",
	"0.00":         "0.00",
}

// Trf is the format string translated into lang by Tr, then filled with args
// the way Sprintf fills one.
func Trf(sandbox *api.Sandbox, lang string, format string, args ...any) string {
	return sandbox.Deps.StdDeps.Sprintf(Tr(sandbox, lang, format), args...)
}

// Translator is Tr bound to lang, the T a page template calls:
// {{call .T "Save"}}.
func Translator(sandbox *api.Sandbox, lang string) func(string) string {
	return func(text string) string {
		return Tr(sandbox, lang, text)
	}
}

// Formatter is Trf bound to lang, the Tf a page template calls:
// {{call .Tf "%d of %d" .From .Total}}.
func Formatter(sandbox *api.Sandbox, lang string) func(string, ...any) string {
	return func(format string, args ...any) string {
		return Trf(sandbox, lang, format, args...)
	}
}
