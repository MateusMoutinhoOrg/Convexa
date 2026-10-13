package erpstore

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
)

// The settings of Convexa are a few named values a root may change: each is
// one record of the setting table, and an unset one reads as its default.

// SettingCurrency is the code of the currency every amount is kept in.
const SettingCurrency = "currency"

// DefaultCurrency is the currency when none was set.
const DefaultCurrency = "BRL"

// Settings is every setting, set or not.
type Settings struct {
	Currency string
}

// ReadSettings is every setting, the unset ones at their defaults.
func ReadSettings(sandbox *api.Sandbox) Settings {
	settings := Settings{Currency: DefaultCurrency}
	if record, ok := db(sandbox).FindSettingByName(SettingCurrency); ok && record.Value != "" {
		settings.Currency = record.Value
	}
	return settings
}

// Currency is the code of the currency every amount is kept in.
func Currency(sandbox *api.Sandbox) string {
	return ReadSettings(sandbox).Currency
}

// WriteSettings writes settings.
func WriteSettings(sandbox *api.Sandbox, settings Settings) (Problem, error) {
	defer lock(sandbox)()
	currency := sandbox.Deps.StringsDeps.ToUpper(clean(sandbox, settings.Currency, 3))
	if len(currency) != 3 {
		return invalid(sandbox, "currency", "Enter the currency as its three-letter code, such as BRL, USD or EUR."), nil
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return invalid(sandbox, "currency", "Enter the currency as its three-letter code, such as BRL, USD or EUR."), nil
		}
	}
	return Problem{}, writeSetting(sandbox, SettingCurrency, currency)
}

// writeSetting writes value under name.
func writeSetting(sandbox *api.Sandbox, name string, value string) error {
	store := db(sandbox)
	if record, ok := store.FindSettingByName(name); ok {
		return store.SetSettingValue(record.Id, value)
	}
	_, err := store.AddSetting(erp_db.SettingInput{Name: name, Value: value})
	return err
}
