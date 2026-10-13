package erpmoney

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
)

// Money is stored and computed as an integer number of cents — 123456 is
// 1,234.56 — never as a float, so a sum of a thousand entries is exact. The
// api reads and answers amounts in cents too; only a page spells them with
// separators, the way its language writes numbers.

// Format is cents as lang writes an amount: 1,234.56 in English, 1.234,56 in
// Portuguese, with a leading minus when negative.
func Format(sandbox *api.Sandbox, lang string, cents int64) string {
	thousands, decimal := ",", "."
	if lang == erpi18n.Portuguese {
		thousands, decimal = ".", ","
	}
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	whole := sandbox.Deps.StringsDeps.FormatInt(cents/100, 10)
	grouped := ""
	for len(whole) > 3 {
		grouped = thousands + whole[len(whole)-3:] + grouped
		whole = whole[:len(whole)-3]
	}
	grouped = whole + grouped
	return sandbox.Deps.StdDeps.Sprintf("%s%s%s%02d", sign, grouped, decimal, cents%100)
}

// Symbol is the sign a currency code is written with, the code itself for
// one it does not know.
func Symbol(sandbox *api.Sandbox, currency string) string {
	switch sandbox.Deps.StringsDeps.ToUpper(currency) {
	case "BRL":
		return "R$"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "¥"
	}
	return currency
}

// FormatCurrency is Format with the currency's symbol in front: R$ 1.234,56.
func FormatCurrency(sandbox *api.Sandbox, lang string, currency string, cents int64) string {
	return Symbol(sandbox, currency) + " " + Format(sandbox, lang, cents)
}

// FormatInput is cents as an amount field is filled with — no thousands
// separator, the decimal one of lang — or "" for 0, so a blank field stays
// blank.
func FormatInput(sandbox *api.Sandbox, lang string, cents int64) string {
	if cents == 0 {
		return ""
	}
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	decimal := "."
	if lang == erpi18n.Portuguese {
		decimal = ","
	}
	return sandbox.Deps.StdDeps.Sprintf("%s%d%s%02d", sign, cents/100, decimal, cents%100)
}

// Parse is the amount text spells, in cents, or false when it spells none.
// It reads what a person types in either language: 1234.56, 1,234.56,
// 1.234,56, 1234,5, R$ 10, -12. The last '.' or ',' is the decimal separator
// when it appears once and one or two digits follow it; every other '.' or ','
// separates thousands — so 1.234 and 1,234 both read as one thousand two
// hundred and thirty-four, since no amount carries three decimals.
func Parse(sandbox *api.Sandbox, text string) (int64, bool) {
	cleaned := []byte{}
	negative := false
	for _, char := range []byte(sandbox.Deps.StringsDeps.TrimSpace(text)) {
		switch {
		case char >= '0' && char <= '9', char == '.', char == ',':
			cleaned = append(cleaned, char)
		case char == '-':
			if len(cleaned) > 0 || negative {
				return 0, false
			}
			negative = true
		case char == ' ', char == 'R', char == '$', char == '+':
		default:
			return 0, false
		}
	}
	if len(cleaned) == 0 {
		return 0, false
	}
	decimalAt := -1
	last := -1
	for i, char := range cleaned {
		if char == '.' || char == ',' {
			last = i
		}
	}
	if last >= 0 {
		trailing := len(cleaned) - last - 1
		same := 0
		for _, char := range cleaned {
			if char == cleaned[last] {
				same++
			}
		}
		if (trailing == 1 || trailing == 2) && same == 1 {
			decimalAt = last
		}
	}
	var whole, fraction int64
	digits := 0
	for i, char := range cleaned {
		if char == '.' || char == ',' {
			continue
		}
		if decimalAt >= 0 && i > decimalAt {
			fraction = fraction*10 + int64(char-'0')
			digits++
			continue
		}
		whole = whole*10 + int64(char-'0')
		if whole > 90_000_000_000_000 {
			return 0, false
		}
	}
	if digits == 1 {
		fraction *= 10
	}
	cents := whole*100 + fraction
	if negative {
		cents = -cents
	}
	return cents, true
}

// Abs is cents without its sign.
func Abs(sandbox *api.Sandbox, cents int64) int64 {
	if cents < 0 {
		return -cents
	}
	return cents
}
