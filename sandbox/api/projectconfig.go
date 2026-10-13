package api

// ProjectConfig is the part of the Config the project declares itself: api.Config
// embeds it, so every field typed here is read as sandbox.Config.<Field>,
// beside the ProjectName and Version the build renders from project.yaml.
//
// Fill it in sandbox/constructors/config/constructor.go, after NewConfig — that
// file is written once and is yours too. A field is written in the builtin
// types or in a type of this package, like every other of sandbox/api/.
//
// Written once by `agnos start` and then yours: no build rewrites
// this file.
type ProjectConfig struct {
	// UtcOffsetMinutes is how far from UTC the dates of Convexa are read —
	// what "today" is, when a payable is overdue — in minutes: -180 for
	// Brasília. The erp-start-server middleware fills it from --utc-offset in
	// front of start-server; 0 elsewhere.
	UtcOffsetMinutes int64

	// RecurrenceHorizonDays is how many days ahead of today the recurring
	// entries are turned into pending transactions, so the payables and
	// receivables list them before they are due. The erp-start-server
	// middleware fills it from --recurrence-horizon; 31 elsewhere.
	RecurrenceHorizonDays int64
}
