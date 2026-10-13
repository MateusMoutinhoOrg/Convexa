package erpdate

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpi18n"
)

// A date is stored and passed around as one integer spelling it YYYYMMDD —
// 20261012 is the 12th of October 2026 — and a month as YYYYMM. The store has
// no date type: an integer sorts the way the calendar does, a range of dates
// is a Min/Max filter, it reads the same on disk as on a page, and it carries
// no time zone to get wrong. 0 stands for "no date", which a filter reads as
// "off" too. The api spells a date the ISO way, "2026-10-12".

// Make is the date of day, month and year, unchecked.
func Make(sandbox *api.Sandbox, year int64, month int64, day int64) int64 {
	return year*10000 + month*100 + day
}

// Split is the year, month and day of date.
func Split(sandbox *api.Sandbox, date int64) (int64, int64, int64) {
	return date / 10000, (date / 100) % 100, date % 100
}

// IsLeap tells whether year has a 29th of February.
func IsLeap(sandbox *api.Sandbox, year int64) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth is how many days month of year has.
func DaysInMonth(sandbox *api.Sandbox, year int64, month int64) int64 {
	switch month {
	case 2:
		if IsLeap(sandbox, year) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	}
	return 31
}

// Valid tells whether date names a day of the calendar, between the years
// 1900 and 9999.
func Valid(sandbox *api.Sandbox, date int64) bool {
	year, month, day := Split(sandbox, date)
	if year < 1900 || year > 9999 || month < 1 || month > 12 || day < 1 {
		return false
	}
	return day <= DaysInMonth(sandbox, year, month)
}

// toDays is how many days date lies after 1970-01-01, negative before it: the
// civil-from-days algorithm run backwards, in integers only.
func toDays(sandbox *api.Sandbox, date int64) int64 {
	year, month, day := Split(sandbox, date)
	if month <= 2 {
		year--
	}
	era := year
	if era < 0 {
		era -= 399
	}
	era /= 400
	yearOfEra := year - era*400
	shifted := month + 9
	if month > 2 {
		shifted = month - 3
	}
	dayOfYear := (153*shifted+2)/5 + day - 1
	dayOfEra := yearOfEra*365 + yearOfEra/4 - yearOfEra/100 + dayOfYear
	return era*146097 + dayOfEra - 719468
}

// fromDays is the date days after 1970-01-01.
func fromDays(sandbox *api.Sandbox, days int64) int64 {
	days += 719468
	era := days
	if era < 0 {
		era -= 146096
	}
	era /= 146097
	dayOfEra := days - era*146097
	yearOfEra := (dayOfEra - dayOfEra/1460 + dayOfEra/36524 - dayOfEra/146096) / 365
	year := yearOfEra + era*400
	dayOfYear := dayOfEra - (365*yearOfEra + yearOfEra/4 - yearOfEra/100)
	shifted := (5*dayOfYear + 2) / 153
	day := dayOfYear - (153*shifted+2)/5 + 1
	month := shifted + 3
	if shifted >= 10 {
		month = shifted - 9
	}
	if month <= 2 {
		year++
	}
	return Make(sandbox, year, month, day)
}

// AddDays is the date days after date, before it when days is negative.
func AddDays(sandbox *api.Sandbox, date int64, days int64) int64 {
	return fromDays(sandbox, toDays(sandbox, date)+days)
}

// DaysBetween is how many days to lies after from, negative when before.
func DaysBetween(sandbox *api.Sandbox, from int64, to int64) int64 {
	return toDays(sandbox, to) - toDays(sandbox, from)
}

// Weekday is the day of the week of date, 0 for Sunday through 6 for Saturday.
func Weekday(sandbox *api.Sandbox, date int64) int64 {
	weekday := (toDays(sandbox, date) + 4) % 7
	if weekday < 0 {
		weekday += 7
	}
	return weekday
}

// AddMonths is the date months after date, before it when months is
// negative, its day clamped to the length of the month it lands in: one
// month after the 31st of January is the 28th — or 29th — of February.
func AddMonths(sandbox *api.Sandbox, date int64, months int64) int64 {
	year, month, day := Split(sandbox, date)
	year, month = shiftMonth(sandbox, year, month, months)
	if last := DaysInMonth(sandbox, year, month); day > last {
		day = last
	}
	return Make(sandbox, year, month, day)
}

// shiftMonth is the year and month months after month of year.
func shiftMonth(sandbox *api.Sandbox, year int64, month int64, months int64) (int64, int64) {
	index := year*12 + (month - 1) + months
	return index / 12, index%12 + 1
}

// Month is the month date lies in, as YYYYMM.
func Month(sandbox *api.Sandbox, date int64) int64 {
	return date / 100
}

// AddToMonth is the month months after month (both YYYYMM).
func AddToMonth(sandbox *api.Sandbox, month int64, months int64) int64 {
	year, shifted := shiftMonth(sandbox, month/100, month%100, months)
	return year*100 + shifted
}

// MonthsBetween is how many months to (YYYYMM) lies after from (YYYYMM).
func MonthsBetween(sandbox *api.Sandbox, from int64, to int64) int64 {
	return (to/100*12 + to%100) - (from/100*12 + from%100)
}

// DayOfMonth is day of month (YYYYMM), clamped to the month's length: the
// 31st of a month of 30 days is its 30th.
func DayOfMonth(sandbox *api.Sandbox, month int64, day int64) int64 {
	year, number := month/100, month%100
	if last := DaysInMonth(sandbox, year, number); day > last {
		day = last
	}
	if day < 1 {
		day = 1
	}
	return Make(sandbox, year, number, day)
}

// MonthStart is the first day of the month date lies in.
func MonthStart(sandbox *api.Sandbox, date int64) int64 {
	return Month(sandbox, date)*100 + 1
}

// MonthEnd is the last day of the month date lies in.
func MonthEnd(sandbox *api.Sandbox, date int64) int64 {
	year, month, _ := Split(sandbox, date)
	return Make(sandbox, year, month, DaysInMonth(sandbox, year, month))
}

// FromUnix is the date the instant seconds falls on, offsetMinutes away from
// UTC: -180 for Brasília.
func FromUnix(sandbox *api.Sandbox, seconds int64, offsetMinutes int64) int64 {
	shifted := seconds + offsetMinutes*60
	days := shifted / 86400
	if shifted < 0 && shifted%86400 != 0 {
		days--
	}
	return fromDays(sandbox, days)
}

// Today is the date it is now, where the server's clock is read
// sandbox.Config.UtcOffsetMinutes away from UTC.
func Today(sandbox *api.Sandbox) int64 {
	return FromUnix(sandbox, sandbox.Deps.StdDeps.Now()/1_000_000_000, sandbox.Config.UtcOffsetMinutes)
}

// ParseISO is the date text spells as 2026-10-12, or false when it spells
// none.
func ParseISO(sandbox *api.Sandbox, text string) (int64, bool) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if len(text) != 10 || text[4] != '-' || text[7] != '-' {
		return 0, false
	}
	year, errYear := sandbox.Deps.StringsDeps.ParseInt(text[0:4], 10, 64)
	month, errMonth := sandbox.Deps.StringsDeps.ParseInt(text[5:7], 10, 64)
	day, errDay := sandbox.Deps.StringsDeps.ParseInt(text[8:10], 10, 64)
	if errYear != nil || errMonth != nil || errDay != nil {
		return 0, false
	}
	date := Make(sandbox, year, month, day)
	if !Valid(sandbox, date) {
		return 0, false
	}
	return date, true
}

// ParseOptionalISO is ParseISO for a field that may be left blank: "" is the
// date 0 and still ok.
func ParseOptionalISO(sandbox *api.Sandbox, text string) (int64, bool) {
	if sandbox.Deps.StringsDeps.TrimSpace(text) == "" {
		return 0, true
	}
	return ParseISO(sandbox, text)
}

// ISO spells date the ISO way, 2026-10-12, or "" for the date 0.
func ISO(sandbox *api.Sandbox, date int64) string {
	if date == 0 {
		return ""
	}
	year, month, day := Split(sandbox, date)
	return sandbox.Deps.StdDeps.Sprintf("%04d-%02d-%02d", year, month, day)
}

// ParseMonth is the month (YYYYMM) text spells as 2026-10, or false.
func ParseMonth(sandbox *api.Sandbox, text string) (int64, bool) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if len(text) != 7 || text[4] != '-' {
		return 0, false
	}
	year, errYear := sandbox.Deps.StringsDeps.ParseInt(text[0:4], 10, 64)
	month, errMonth := sandbox.Deps.StringsDeps.ParseInt(text[5:7], 10, 64)
	if errYear != nil || errMonth != nil || month < 1 || month > 12 || year < 1900 || year > 9999 {
		return 0, false
	}
	return year*100 + month, true
}

// ISOMonth spells month (YYYYMM) as 2026-10, or "" for 0.
func ISOMonth(sandbox *api.Sandbox, month int64) string {
	if month == 0 {
		return ""
	}
	return sandbox.Deps.StdDeps.Sprintf("%04d-%02d", month/100, month%100)
}

// monthNames are the short names of the months, in English and in Portuguese.
var monthNames = map[string][]string{
	erpi18n.English:    {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	erpi18n.Portuguese: {"jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"},
}

// MonthName is the short name of the month number (1-12) in lang.
func MonthName(sandbox *api.Sandbox, lang string, month int64) string {
	names, ok := monthNames[lang]
	if !ok {
		names = monthNames[erpi18n.English]
	}
	if month < 1 || month > 12 {
		return ""
	}
	return names[month-1]
}

// Format is date as a person reads it in lang — Oct 12, 2026 in English,
// 12/10/2026 in Portuguese — or "" for the date 0.
func Format(sandbox *api.Sandbox, lang string, date int64) string {
	if date == 0 {
		return ""
	}
	year, month, day := Split(sandbox, date)
	if lang == erpi18n.Portuguese {
		return sandbox.Deps.StdDeps.Sprintf("%02d/%02d/%04d", day, month, year)
	}
	return sandbox.Deps.StdDeps.Sprintf("%s %d, %04d", MonthName(sandbox, lang, month), day, year)
}

// FormatMonth is month (YYYYMM) as a person reads it in lang — Oct 2026, or
// out/2026 — or "" for 0.
func FormatMonth(sandbox *api.Sandbox, lang string, month int64) string {
	if month == 0 {
		return ""
	}
	if lang == erpi18n.Portuguese {
		return sandbox.Deps.StdDeps.Sprintf("%s/%04d", MonthName(sandbox, lang, month%100), month/100)
	}
	return sandbox.Deps.StdDeps.Sprintf("%s %04d", MonthName(sandbox, lang, month%100), month/100)
}

// FormatInstant is the instant seconds as a person reads it in lang, in the
// server's offset from UTC.
func FormatInstant(sandbox *api.Sandbox, lang string, seconds int64) string {
	if seconds == 0 {
		return ""
	}
	shifted := seconds + sandbox.Config.UtcOffsetMinutes*60
	layout := "Jan 2, 2006 15:04"
	if lang == erpi18n.Portuguese {
		layout = "02/01/2006 15:04"
	}
	return sandbox.Deps.TimeDeps.FormatUnix(shifted, layout)
}
