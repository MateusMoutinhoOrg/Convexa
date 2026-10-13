package erpreports

import (
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/api"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/databases/erp_db"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpdate"
	"github.com/MateusMoutinhoOrg/Convexa/sandbox/internal/erp/erpledger"
)

// AccountBalance is the balance of one account at a date: the whole of it —
// the cash it holds, or the debt of a card — and the share of each company.
type AccountBalance struct {
	Account erp_db.AccountRecord
	// Total is the absolute balance: what the account holds, every company
	// counted.
	Total int64
	// ByCompany is the share of each company, by company id; a company with
	// no transaction on the account is absent.
	ByCompany map[int64]int64
	// Pending is what the pending transactions will move, every date counted.
	Pending int64
}

// Share is the balance of the company with id company on balance.
func Share(sandbox *api.Sandbox, balance AccountBalance, company int64) int64 {
	return balance.ByCompany[company]
}

// AccountBalances is the balance at date of every account — active ones, and
// archived ones still holding something — by name.
func AccountBalances(sandbox *api.Sandbox, book Book, date int64) []AccountBalance {
	byAccount := map[int64]*AccountBalance{}
	for _, account := range book.Accounts {
		byAccount[account.Id] = &AccountBalance{Account: account, ByCompany: map[int64]int64{}}
	}
	for _, tx := range book.Transactions {
		balance, ok := byAccount[tx.Account]
		if !ok {
			continue
		}
		if erpledger.SettledBy(sandbox, tx, date) {
			balance.Total += tx.Amount
			balance.ByCompany[tx.Company] += tx.Amount
		} else if tx.Status == erpledger.StatusPending {
			balance.Pending += tx.Amount
			if _, seen := balance.ByCompany[tx.Company]; !seen {
				balance.ByCompany[tx.Company] = 0
			}
		}
	}
	balances := []AccountBalance{}
	for _, account := range book.Accounts {
		balance := byAccount[account.Id]
		if account.Status != "active" && balance.Total == 0 && balance.Pending == 0 {
			continue
		}
		balances = append(balances, *balance)
	}
	return balances
}

// BalanceTotals is the sum of balances over the cash accounts — what the
// companies hold — and over the cards — what they owe, positive — with the
// share of each company of both.
type BalanceTotals struct {
	Cash          int64
	CardDebt      int64
	CashByCompany map[int64]int64
	DebtByCompany map[int64]int64
}

// TotalsOf is the BalanceTotals of balances.
func TotalsOf(sandbox *api.Sandbox, balances []AccountBalance) BalanceTotals {
	totals := BalanceTotals{CashByCompany: map[int64]int64{}, DebtByCompany: map[int64]int64{}}
	for _, balance := range balances {
		card := erpledger.IsCard(sandbox, balance.Account)
		for company, share := range balance.ByCompany {
			if card {
				totals.DebtByCompany[company] -= share
			} else {
				totals.CashByCompany[company] += share
			}
		}
		if card {
			totals.CardDebt -= balance.Total
		} else {
			totals.Cash += balance.Total
		}
	}
	return totals
}

// StatementLine is one settled transaction of a statement and the balance
// right after it.
type StatementLine struct {
	Tx      erp_db.TransactionRecord
	Balance int64
}

// Statement is what an account — one company's share of it, or the whole —
// went through over a period: the transactions settled in it, in order, with
// the running balance.
type Statement struct {
	Account  erp_db.AccountRecord
	Company  int64
	From     int64
	To       int64
	Opening  int64
	Inflows  int64
	Outflows int64
	Closing  int64
	Lines    []StatementLine
}

// AccountStatement is the statement of the account with id account — of the
// company with id company alone, unless 0 — between from and to, both
// included, by settlement date.
func AccountStatement(sandbox *api.Sandbox, book Book, account int64, company int64, from int64, to int64) Statement {
	statement := Statement{Account: book.AccountById[account], Company: company, From: from, To: to}
	statement.Opening = erpledger.Balance(sandbox, book.Transactions, account, company, erpdate.AddDays(sandbox, from, -1))
	settled := []erp_db.TransactionRecord{}
	for _, tx := range book.Transactions {
		if tx.Account != account || (company != 0 && tx.Company != company) || tx.Status != erpledger.StatusSettled {
			continue
		}
		if tx.SettledDate < from || tx.SettledDate > to {
			continue
		}
		settled = append(settled, tx)
	}
	sandbox.Deps.SortDeps.Slice(settled, func(i int, j int) bool {
		if settled[i].SettledDate != settled[j].SettledDate {
			return settled[i].SettledDate < settled[j].SettledDate
		}
		return settled[i].Id < settled[j].Id
	})
	running := statement.Opening
	for _, tx := range settled {
		running += tx.Amount
		if tx.Amount >= 0 {
			statement.Inflows += tx.Amount
		} else {
			statement.Outflows -= tx.Amount
		}
		statement.Lines = append(statement.Lines, StatementLine{Tx: tx, Balance: running})
	}
	statement.Closing = running
	return statement
}
