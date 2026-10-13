# `ErpDb`

`sandbox/internal/databases/erp_db/`, keys under `erpdb`. Build one with
`erp_db.New(sandbox)` — it touches no key, so building one is free.

## `company`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `string` | yes |  |
| `document` | `string` |  |  |
| `notes` | `string` |  |  |
| `status` | `string` | yes |  |
| `created-at` | `integer` |  |  |

## `account`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `string` | yes |  |
| `kind` | `string` | yes |  |
| `institution` | `string` |  |  |
| `notes` | `string` |  |  |
| `status` | `string` | yes |  |
| `credit-limit` | `integer` |  |  |
| `closing-day` | `integer` |  |  |
| `due-day` | `integer` |  |  |
| `payment-account` | `link` |  | `account` |
| `created-at` | `integer` |  |  |

## `category`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `string` | yes |  |
| `type` | `string` | yes |  |
| `parent` | `link` |  | `category` |
| `status` | `string` | yes |  |
| `created-at` | `integer` |  |  |

## `third-party`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `string` | yes |  |
| `role` | `string` | yes |  |
| `document` | `string` |  |  |
| `email` | `string` |  |  |
| `phone` | `string` |  |  |
| `address` | `string` |  |  |
| `notes` | `string` |  |  |
| `status` | `string` | yes |  |
| `created-at` | `integer` |  |  |

## `recurrence`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `company` | `link` | yes | `company` |
| `account` | `link` | yes | `account` |
| `category` | `link` |  | `category` |
| `third-party` | `link` |  | `third-party` |
| `type` | `string` | yes |  |
| `amount` | `integer` | yes |  |
| `description` | `string` | yes |  |
| `notes` | `string` |  |  |
| `frequency` | `string` | yes |  |
| `interval` | `integer` | yes |  |
| `start-date` | `integer` | yes |  |
| `end-date` | `integer` |  |  |
| `generated-until` | `integer` |  |  |
| `status` | `string` | yes |  |
| `created-at` | `integer` |  |  |

## `transaction`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `company` | `link` | yes | `company` |
| `account` | `link` | yes | `account` |
| `category` | `link` |  | `category` |
| `third-party` | `link` |  | `third-party` |
| `type` | `string` | yes |  |
| `amount` | `integer` | yes |  |
| `description` | `string` | yes |  |
| `document-number` | `string` |  |  |
| `notes` | `string` |  |  |
| `date` | `integer` | yes |  |
| `due-date` | `integer` |  |  |
| `settled-date` | `integer` |  |  |
| `status` | `string` | yes |  |
| `counterpart` | `link` |  | `transaction` |
| `installment-group` | `integer` |  |  |
| `installment-number` | `integer` |  |  |
| `installment-count` | `integer` |  |  |
| `recurrence` | `link` |  | `recurrence` |
| `occurrence` | `integer` |  |  |
| `card-invoice` | `integer` |  |  |
| `created-at` | `integer` |  |  |
| `updated-at` | `integer` |  |  |
| `created-by` | `integer` |  |  |

## `budget`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `company` | `link` | yes | `company` |
| `category` | `link` | yes | `category` |
| `period-start` | `integer` | yes |  |
| `period-end` | `integer` | yes |  |
| `amount` | `integer` | yes |  |
| `notes` | `string` |  |  |
| `created-at` | `integer` |  |  |

## `setting`

| Field | Type | Required | Target |
| --- | --- | --- | --- |
| `name` | `key` | yes |  |
| `value` | `string` |  |  |

## Methods

| Method | What it does |
| --- | --- |
| `AddCompany(props CompanyInput) (CompanyRecord, error)` | inserts one company record |
| `FindCompanyById(id int64) (CompanyRecord, bool)` | reads one company record by its permanent id |
| `ListCompanies(filter CompanyFilter) ([]CompanyRecord, error)` | reads every company record the filter keeps |
| `ListCompaniesPage(offset int, limit int) ([]CompanyRecord, error)` | reads up to limit company records after the first offset, every one past them when limit is 0 |
| `CountCompany() (int, error)` | is how many company records are live |
| `SetCompanyName(id int64, value string) error` | writes a new name on one company record |
| `SetCompanyDocument(id int64, value string) error` | writes a new document on one company record |
| `SetCompanyNotes(id int64, value string) error` | writes a new notes on one company record |
| `SetCompanyStatus(id int64, value string) error` | writes a new status on one company record |
| `SetCompanyCreatedAt(id int64, value int64) error` | writes a new created-at on one company record |
| `RemoveCompany(id int64) error` | deletes one company record and everything nested under it |
| `AddAccount(props AccountInput) (AccountRecord, error)` | inserts one account record |
| `FindAccountById(id int64) (AccountRecord, bool)` | reads one account record by its permanent id |
| `ListAccounts(filter AccountFilter) ([]AccountRecord, error)` | reads every account record the filter keeps |
| `ListAccountsPage(offset int, limit int) ([]AccountRecord, error)` | reads up to limit account records after the first offset, every one past them when limit is 0 |
| `CountAccount() (int, error)` | is how many account records are live |
| `SetAccountName(id int64, value string) error` | writes a new name on one account record |
| `SetAccountKind(id int64, value string) error` | writes a new kind on one account record |
| `SetAccountInstitution(id int64, value string) error` | writes a new institution on one account record |
| `SetAccountNotes(id int64, value string) error` | writes a new notes on one account record |
| `SetAccountStatus(id int64, value string) error` | writes a new status on one account record |
| `SetAccountCreditLimit(id int64, value int64) error` | writes a new credit-limit on one account record |
| `SetAccountClosingDay(id int64, value int64) error` | writes a new closing-day on one account record |
| `SetAccountDueDay(id int64, value int64) error` | writes a new due-day on one account record |
| `SetAccountPaymentAccount(id int64, value int64) error` | writes a new payment-account on one account record |
| `SetAccountCreatedAt(id int64, value int64) error` | writes a new created-at on one account record |
| `RemoveAccount(id int64) error` | deletes one account record and everything nested under it |
| `GetAccountPaymentAccount(id int64) (AccountRecord, bool)` | resolves the payment-account link of one account record to the account it points at |
| `AddCategory(props CategoryInput) (CategoryRecord, error)` | inserts one category record |
| `FindCategoryById(id int64) (CategoryRecord, bool)` | reads one category record by its permanent id |
| `ListCategories(filter CategoryFilter) ([]CategoryRecord, error)` | reads every category record the filter keeps |
| `ListCategoriesPage(offset int, limit int) ([]CategoryRecord, error)` | reads up to limit category records after the first offset, every one past them when limit is 0 |
| `CountCategory() (int, error)` | is how many category records are live |
| `SetCategoryName(id int64, value string) error` | writes a new name on one category record |
| `SetCategoryType(id int64, value string) error` | writes a new type on one category record |
| `SetCategoryParent(id int64, value int64) error` | writes a new parent on one category record |
| `SetCategoryStatus(id int64, value string) error` | writes a new status on one category record |
| `SetCategoryCreatedAt(id int64, value int64) error` | writes a new created-at on one category record |
| `RemoveCategory(id int64) error` | deletes one category record and everything nested under it |
| `GetCategoryParent(id int64) (CategoryRecord, bool)` | resolves the parent link of one category record to the category it points at |
| `AddThirdParty(props ThirdPartyInput) (ThirdPartyRecord, error)` | inserts one third-party record |
| `FindThirdPartyById(id int64) (ThirdPartyRecord, bool)` | reads one third-party record by its permanent id |
| `ListThirdParties(filter ThirdPartyFilter) ([]ThirdPartyRecord, error)` | reads every third-party record the filter keeps |
| `ListThirdPartiesPage(offset int, limit int) ([]ThirdPartyRecord, error)` | reads up to limit third-party records after the first offset, every one past them when limit is 0 |
| `CountThirdParty() (int, error)` | is how many third-party records are live |
| `SetThirdPartyName(id int64, value string) error` | writes a new name on one third-party record |
| `SetThirdPartyRole(id int64, value string) error` | writes a new role on one third-party record |
| `SetThirdPartyDocument(id int64, value string) error` | writes a new document on one third-party record |
| `SetThirdPartyEmail(id int64, value string) error` | writes a new email on one third-party record |
| `SetThirdPartyPhone(id int64, value string) error` | writes a new phone on one third-party record |
| `SetThirdPartyAddress(id int64, value string) error` | writes a new address on one third-party record |
| `SetThirdPartyNotes(id int64, value string) error` | writes a new notes on one third-party record |
| `SetThirdPartyStatus(id int64, value string) error` | writes a new status on one third-party record |
| `SetThirdPartyCreatedAt(id int64, value int64) error` | writes a new created-at on one third-party record |
| `RemoveThirdParty(id int64) error` | deletes one third-party record and everything nested under it |
| `AddRecurrence(props RecurrenceInput) (RecurrenceRecord, error)` | inserts one recurrence record |
| `FindRecurrenceById(id int64) (RecurrenceRecord, bool)` | reads one recurrence record by its permanent id |
| `ListRecurrences(filter RecurrenceFilter) ([]RecurrenceRecord, error)` | reads every recurrence record the filter keeps |
| `ListRecurrencesPage(offset int, limit int) ([]RecurrenceRecord, error)` | reads up to limit recurrence records after the first offset, every one past them when limit is 0 |
| `CountRecurrence() (int, error)` | is how many recurrence records are live |
| `SetRecurrenceCompany(id int64, value int64) error` | writes a new company on one recurrence record |
| `SetRecurrenceAccount(id int64, value int64) error` | writes a new account on one recurrence record |
| `SetRecurrenceCategory(id int64, value int64) error` | writes a new category on one recurrence record |
| `SetRecurrenceThirdParty(id int64, value int64) error` | writes a new third-party on one recurrence record |
| `SetRecurrenceType(id int64, value string) error` | writes a new type on one recurrence record |
| `SetRecurrenceAmount(id int64, value int64) error` | writes a new amount on one recurrence record |
| `SetRecurrenceDescription(id int64, value string) error` | writes a new description on one recurrence record |
| `SetRecurrenceNotes(id int64, value string) error` | writes a new notes on one recurrence record |
| `SetRecurrenceFrequency(id int64, value string) error` | writes a new frequency on one recurrence record |
| `SetRecurrenceInterval(id int64, value int64) error` | writes a new interval on one recurrence record |
| `SetRecurrenceStartDate(id int64, value int64) error` | writes a new start-date on one recurrence record |
| `SetRecurrenceEndDate(id int64, value int64) error` | writes a new end-date on one recurrence record |
| `SetRecurrenceGeneratedUntil(id int64, value int64) error` | writes a new generated-until on one recurrence record |
| `SetRecurrenceStatus(id int64, value string) error` | writes a new status on one recurrence record |
| `SetRecurrenceCreatedAt(id int64, value int64) error` | writes a new created-at on one recurrence record |
| `RemoveRecurrence(id int64) error` | deletes one recurrence record and everything nested under it |
| `GetRecurrenceCompany(id int64) (CompanyRecord, bool)` | resolves the company link of one recurrence record to the company it points at |
| `GetRecurrenceAccount(id int64) (AccountRecord, bool)` | resolves the account link of one recurrence record to the account it points at |
| `GetRecurrenceCategory(id int64) (CategoryRecord, bool)` | resolves the category link of one recurrence record to the category it points at |
| `GetRecurrenceThirdParty(id int64) (ThirdPartyRecord, bool)` | resolves the third-party link of one recurrence record to the third-party it points at |
| `AddTransaction(props TransactionInput) (TransactionRecord, error)` | inserts one transaction record |
| `FindTransactionById(id int64) (TransactionRecord, bool)` | reads one transaction record by its permanent id |
| `ListTransactions(filter TransactionFilter) ([]TransactionRecord, error)` | reads every transaction record the filter keeps |
| `ListTransactionsPage(offset int, limit int) ([]TransactionRecord, error)` | reads up to limit transaction records after the first offset, every one past them when limit is 0 |
| `CountTransaction() (int, error)` | is how many transaction records are live |
| `SetTransactionCompany(id int64, value int64) error` | writes a new company on one transaction record |
| `SetTransactionAccount(id int64, value int64) error` | writes a new account on one transaction record |
| `SetTransactionCategory(id int64, value int64) error` | writes a new category on one transaction record |
| `SetTransactionThirdParty(id int64, value int64) error` | writes a new third-party on one transaction record |
| `SetTransactionType(id int64, value string) error` | writes a new type on one transaction record |
| `SetTransactionAmount(id int64, value int64) error` | writes a new amount on one transaction record |
| `SetTransactionDescription(id int64, value string) error` | writes a new description on one transaction record |
| `SetTransactionDocumentNumber(id int64, value string) error` | writes a new document-number on one transaction record |
| `SetTransactionNotes(id int64, value string) error` | writes a new notes on one transaction record |
| `SetTransactionDate(id int64, value int64) error` | writes a new date on one transaction record |
| `SetTransactionDueDate(id int64, value int64) error` | writes a new due-date on one transaction record |
| `SetTransactionSettledDate(id int64, value int64) error` | writes a new settled-date on one transaction record |
| `SetTransactionStatus(id int64, value string) error` | writes a new status on one transaction record |
| `SetTransactionCounterpart(id int64, value int64) error` | writes a new counterpart on one transaction record |
| `SetTransactionInstallmentGroup(id int64, value int64) error` | writes a new installment-group on one transaction record |
| `SetTransactionInstallmentNumber(id int64, value int64) error` | writes a new installment-number on one transaction record |
| `SetTransactionInstallmentCount(id int64, value int64) error` | writes a new installment-count on one transaction record |
| `SetTransactionRecurrence(id int64, value int64) error` | writes a new recurrence on one transaction record |
| `SetTransactionOccurrence(id int64, value int64) error` | writes a new occurrence on one transaction record |
| `SetTransactionCardInvoice(id int64, value int64) error` | writes a new card-invoice on one transaction record |
| `SetTransactionCreatedAt(id int64, value int64) error` | writes a new created-at on one transaction record |
| `SetTransactionUpdatedAt(id int64, value int64) error` | writes a new updated-at on one transaction record |
| `SetTransactionCreatedBy(id int64, value int64) error` | writes a new created-by on one transaction record |
| `RemoveTransaction(id int64) error` | deletes one transaction record and everything nested under it |
| `GetTransactionCompany(id int64) (CompanyRecord, bool)` | resolves the company link of one transaction record to the company it points at |
| `GetTransactionAccount(id int64) (AccountRecord, bool)` | resolves the account link of one transaction record to the account it points at |
| `GetTransactionCategory(id int64) (CategoryRecord, bool)` | resolves the category link of one transaction record to the category it points at |
| `GetTransactionThirdParty(id int64) (ThirdPartyRecord, bool)` | resolves the third-party link of one transaction record to the third-party it points at |
| `GetTransactionCounterpart(id int64) (TransactionRecord, bool)` | resolves the counterpart link of one transaction record to the transaction it points at |
| `GetTransactionRecurrence(id int64) (RecurrenceRecord, bool)` | resolves the recurrence link of one transaction record to the recurrence it points at |
| `AddBudget(props BudgetInput) (BudgetRecord, error)` | inserts one budget record |
| `FindBudgetById(id int64) (BudgetRecord, bool)` | reads one budget record by its permanent id |
| `ListBudgets(filter BudgetFilter) ([]BudgetRecord, error)` | reads every budget record the filter keeps |
| `ListBudgetsPage(offset int, limit int) ([]BudgetRecord, error)` | reads up to limit budget records after the first offset, every one past them when limit is 0 |
| `CountBudget() (int, error)` | is how many budget records are live |
| `SetBudgetCompany(id int64, value int64) error` | writes a new company on one budget record |
| `SetBudgetCategory(id int64, value int64) error` | writes a new category on one budget record |
| `SetBudgetPeriodStart(id int64, value int64) error` | writes a new period-start on one budget record |
| `SetBudgetPeriodEnd(id int64, value int64) error` | writes a new period-end on one budget record |
| `SetBudgetAmount(id int64, value int64) error` | writes a new amount on one budget record |
| `SetBudgetNotes(id int64, value string) error` | writes a new notes on one budget record |
| `SetBudgetCreatedAt(id int64, value int64) error` | writes a new created-at on one budget record |
| `RemoveBudget(id int64) error` | deletes one budget record and everything nested under it |
| `GetBudgetCompany(id int64) (CompanyRecord, bool)` | resolves the company link of one budget record to the company it points at |
| `GetBudgetCategory(id int64) (CategoryRecord, bool)` | resolves the category link of one budget record to the category it points at |
| `AddSetting(props SettingInput) (SettingRecord, error)` | inserts one setting record |
| `FindSettingById(id int64) (SettingRecord, bool)` | reads one setting record by its permanent id |
| `FindSettingByName(value string) (SettingRecord, bool)` | reads one setting record by its indexed name |
| `ListSettings(filter SettingFilter) ([]SettingRecord, error)` | reads every setting record the filter keeps |
| `ListSettingsPage(offset int, limit int) ([]SettingRecord, error)` | reads up to limit setting records after the first offset, every one past them when limit is 0 |
| `CountSetting() (int, error)` | is how many setting records are live |
| `SetSettingName(id int64, value string) error` | writes a new name on one setting record |
| `SetSettingValue(id int64, value string) error` | writes a new value on one setting record |
| `RemoveSetting(id int64) error` | deletes one setting record and everything nested under it |

A query this page does not list goes in `methods_custom.go`, hand-written beside these and
rewritten by no build.

[every database](doc.md)
