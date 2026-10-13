#!/usr/bin/env python3
"""End-to-end smoke test of Convexa.

Builds the binary (unless --binary is given), starts it on a temporary
database folder, creates a root, an admin and a basic user, signs each one in
and creates an API token for each, then drives the ERP through its JSON api
and its pages, checking the numbers that matter: shared accounts, transfers,
credit card installments and invoices, recurrences, budgets and the
projection, the roles and the Portuguese messages.

    python3 scripts/smoke_test.py            # build and test
    python3 scripts/smoke_test.py --keep     # keep the temporary folder

Only the Python standard library is used.
"""
import argparse, http.cookiejar, json, os, re, shutil, socket, subprocess, sys, tempfile, time, urllib.error, urllib.parse, urllib.request

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SECRET = "smoke-test-secret-0123456789abcdefghijklmnop"
failures = []


def check(name, condition, detail=""):
    print(("PASS  " if condition else "FAIL  ") + name + ("" if condition else "  " + str(detail)))
    if not condition:
        failures.append(name)


def free_port():
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


class Client:
    """A browser-like client: keeps the session cookie, posts forms."""

    def __init__(self, base):
        self.base = base
        self.jar = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(self.jar), NoRedirect())

    def request(self, method, path, data=None, headers=None):
        headers = dict(headers or {})
        body = None
        if data is not None:
            body = urllib.parse.urlencode(data).encode()
            headers["Content-Type"] = "application/x-www-form-urlencoded"
            headers["Origin"] = self.base
        req = urllib.request.Request(self.base + path, data=body, method=method, headers=headers)
        try:
            with self.opener.open(req) as response:
                return response.status, response.read().decode(), dict(response.headers)
        except urllib.error.HTTPError as error:
            return error.code, error.read().decode(), dict(error.headers)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *args, **kwargs):
        return None


def api(base, token, path, body=None, lang=None):
    headers = {"Authorization": "Bearer " + token, "Content-Type": "application/json"}
    if lang:
        headers["Accept-Language"] = lang
    req = urllib.request.Request(base + path, data=json.dumps(body or {}).encode(), method="POST", headers=headers)
    try:
        with urllib.request.urlopen(req) as response:
            return response.status, json.loads(response.read().decode() or "{}")
    except urllib.error.HTTPError as error:
        text = error.read().decode()
        try:
            return error.code, json.loads(text)
        except ValueError:
            return error.code, {"raw": text}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--binary", help="a built convexa binary; built from the repo when absent")
    parser.add_argument("--keep", action="store_true", help="keep the temporary folder")
    args = parser.parse_args()

    work = tempfile.mkdtemp(prefix="convexa-smoke-")
    binary = args.binary
    if not binary:
        binary = os.path.join(work, "convexa")
        subprocess.run(["go", "build", "-o", binary, "./cmd/main"], cwd=ROOT, check=True)
    env = dict(os.environ, CONVEXA_BACKOFFICE_SECRET=SECRET)

    passwords = {}
    for role in ["root", "admin", "basic"]:
        out = subprocess.run([binary, "add-backoffice-user", "--username", role + "user", "--email", role + "@example.com", "--role", role, "--database", "data"],
                             cwd=work, env=env, capture_output=True, text=True, check=True).stdout
        passwords[role] = re.search(r"password: (\S+)", out).group(1)

    port = free_port()
    base = "http://127.0.0.1:%d" % port
    server = subprocess.Popen([binary, "start-server", "--insecure-http", "--addr", "127.0.0.1:%d" % port, "--database", "data"],
                              cwd=work, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(100):
            try:
                urllib.request.urlopen(base + "/health")
                break
            except Exception:
                time.sleep(0.1)
        run(base, passwords)
    finally:
        server.terminate()
        server.wait()
        if args.keep:
            print("kept", work)
        else:
            shutil.rmtree(work, ignore_errors=True)
    print()
    if failures:
        print("%d check(s) failed: %s" % (len(failures), ", ".join(failures)))
        sys.exit(1)
    print("every check passed")


def sign_in(base, role, password):
    client = Client(base)
    status, _, headers = client.request("POST", "/admin/login", {"username": role + "user", "password": password})
    check("sign in as " + role, status == 303, status)
    status, page, _ = client.request("POST", "/admin/add-backoffice-api-token", {"name": "smoke", "expiration": "30", "password": password})
    token = re.search(r'value="(bo_[0-9a-f]+)"', page)
    check("API token for " + role, token is not None, status)
    return client, token.group(1) if token else ""


def run(base, passwords):
    root_client, root = sign_in(base, "root", passwords["root"])
    admin_client, admin = sign_in(base, "admin", passwords["admin"])
    basic_client, basic = sign_in(base, "basic", passwords["basic"])

    status, settings = api(base, root, "/api/admin/erp/get-settings")
    today = settings.get("today", "")
    check("settings answer today", status == 200 and len(today) == 10, settings)
    year, month, day = (int(x) for x in today.split("-"))

    def ok(response, expected=(200, 201)):
        status, body = response
        if status not in expected:
            print("      ", status, body)
        return body

    # Companies, accounts, categories, a third party.
    personal = ok(api(base, root, "/api/admin/manage/add-company", {"name": "Personal"}))["company"]["id"]
    acme = ok(api(base, admin, "/api/admin/manage/add-company", {"name": "ACME Ltd", "document": "12.345.678/0001-90"}))["company"]["id"]
    checking = ok(api(base, root, "/api/admin/manage/add-account", {"name": "Main checking", "kind": "checking", "opening-balance": 100000, "opening-company": personal, "opening-date": today}))["account"]["id"]
    savings = ok(api(base, root, "/api/admin/manage/add-account", {"name": "Savings", "kind": "savings"}))["account"]["id"]
    card = ok(api(base, root, "/api/admin/manage/add-account", {"name": "Visa", "kind": "credit-card", "credit-limit": 500000, "closing-day": 3, "due-day": 10, "payment-account": checking}))["account"]["id"]
    salary = ok(api(base, root, "/api/admin/manage/add-category", {"name": "Salary", "type": "income"}))["category"]["id"]
    groceries = ok(api(base, root, "/api/admin/manage/add-category", {"name": "Groceries", "type": "expense"}))["category"]["id"]
    rent = ok(api(base, root, "/api/admin/manage/add-category", {"name": "Rent", "type": "expense"}))["category"]["id"]
    supplier = ok(api(base, root, "/api/admin/manage/add-third-party", {"name": "Market", "role": "supplier"}))["third-party"]["id"]

    # Two companies share the checking account.
    ok(api(base, root, "/api/admin/manage/add-transaction", {"company": acme, "account": checking, "type": "adjustment", "amount": 50000, "date": today, "status": "settled"}))
    ok(api(base, root, "/api/admin/manage/add-transaction", {"company": personal, "account": checking, "category": salary, "type": "income", "amount": 300000, "description": "Salary", "date": today, "status": "settled"}))
    ok(api(base, root, "/api/admin/manage/add-transaction", {"company": acme, "account": checking, "category": groceries, "third-party": supplier, "type": "expense", "amount": 20000, "description": "Supplies", "date": today, "status": "settled"}))
    balances = ok(api(base, root, "/api/admin/erp/account-balances", {}))
    main = next(a for a in balances["accounts"] if a["id"] == checking)
    shares = {b["company"]: b["balance"] for b in main["balances"]}
    check("shared account: absolute balance", main["balance"] == 100000 + 50000 + 300000 - 20000, main["balance"])
    check("shared account: absolute = sum of the companies' shares", main["balance"] == sum(shares.values()), shares)
    check("shared account: each company's share", shares.get(personal) == 400000 and shares.get(acme) == 30000, shares)

    # A transfer moves money without changing the cash.
    cash_before = balances["cash"]
    ok(api(base, root, "/api/admin/manage/add-transfer", {"from-account": checking, "to-account": savings, "from-company": personal, "amount": 10000, "date": today, "status": "settled"}))
    balances = ok(api(base, root, "/api/admin/erp/account-balances", {}))
    by_id = {a["id"]: a for a in balances["accounts"]}
    check("transfer: cash unchanged", balances["cash"] == cash_before, (cash_before, balances["cash"]))
    check("transfer: savings received it", by_id[savings]["balance"] == 10000, by_id[savings]["balance"])

    # A purchase in ten installments lands on ten consecutive invoices.
    created = ok(api(base, root, "/api/admin/manage/add-transaction", {"company": personal, "account": card, "category": groceries, "type": "expense", "amount": 120000, "description": "TV", "date": today, "status": "settled", "installments": 10}))
    invoices = [t["card-invoice"] for t in created.get("transactions", [])]
    months = sorted(set(invoices))
    check("installments: ten transactions", len(invoices) == 10, invoices)
    check("installments: ten consecutive invoices", len(months) == 10, months)
    listing = ok(api(base, root, "/api/admin/erp/list-card-invoices", {"id": card}))
    first = listing["invoices"][0]
    check("invoice: charges of one installment", first["charges"] == 12000, first)
    before = by_id[checking]["balance"]
    paid = ok(api(base, root, "/api/admin/manage/pay-card-invoice", {"id": card, "month": first["month"], "date": today, "status": "settled"}))
    check("invoice: paid what was left", paid.get("paid") == 12000, paid)
    invoice = ok(api(base, root, "/api/admin/erp/get-card-invoice", {"id": card, "month": first["month"]}))["invoice"]
    check("invoice: nothing left after paying", invoice["remaining"] == 0 and invoice["status"] == "paid", invoice)
    balances = ok(api(base, root, "/api/admin/erp/account-balances", {}))
    by_id = {a["id"]: a for a in balances["accounts"]}
    check("invoice: the payment left the checking account", by_id[checking]["balance"] == before - 12000, (before, by_id[checking]["balance"]))

    # A monthly recurrence writes its next occurrence and counts in the projection.
    start = "%04d-%02d-%02d" % (year, month, min(day, 28))
    recurrence = ok(api(base, root, "/api/admin/manage/add-recurrence", {"company": personal, "account": checking, "category": rent, "type": "expense", "amount": 5000, "description": "Rent", "frequency": "monthly", "interval": 1, "start-date": start}))["recurrence"]
    written = ok(api(base, root, "/api/admin/erp/list-transactions", {"recurrence": recurrence["id"]}))
    check("recurrence: occurrences written as transactions", written.get("total", 0) >= 1, written.get("total"))
    projection = ok(api(base, root, "/api/admin/erp/projection", {"months": 12, "events": True}))
    sources = set(e["source"] for e in projection.get("events", []))
    check("projection: occurrences beyond the horizon count", "recurrence" in sources, sources)
    check("projection: the card's future invoices count", "invoice" in sources, sources)

    # A budget lowers the projection by what is left of it.
    ok(api(base, root, "/api/admin/manage/add-budget", {"company": personal, "category": groceries, "period-start": today, "months": 3, "amount": 100000}))
    projection = ok(api(base, root, "/api/admin/erp/projection", {"months": 12}))
    check("budgets: projection with budgets <= without", projection["final-with-budgets"] <= projection["final"], projection)
    check("budgets: something is left of them", projection.get("budgets", 0) < 0, projection.get("budgets"))
    report = ok(api(base, root, "/api/admin/erp/budget-report", {"company": personal}))
    check("budget report lists the budgets", len(report.get("budgets", [])) >= 3, report)

    # Reports answer.
    for path, body in [("/api/admin/erp/dashboard", {}), ("/api/admin/erp/cash-flow", {"basis": "projected", "group": "month"}),
                       ("/api/admin/erp/payables-receivables", {}), ("/api/admin/erp/company-report", {"company": personal}),
                       ("/api/admin/erp/account-statement", {"account": checking}), ("/api/admin/erp/list-credit-cards", {})]:
        status, _ = api(base, basic, path, body)
        check("basic reads " + path, status == 200, status)

    # Roles.
    status, _ = api(base, basic, "/api/admin/manage/add-company", {"name": "Nope"})
    check("basic cannot write", status == 403, status)
    status, _ = api(base, basic, "/api/admin/erp/list-companies", {})
    check("basic can read", status == 200, status)
    req = urllib.request.Request(base + "/api/admin/root/list-backups", headers={"Authorization": "Bearer " + admin})
    try:
        status = urllib.request.urlopen(req).status
    except urllib.error.HTTPError as error:
        status = error.code
    check("admin cannot reach root", status == 403, status)
    status, _, _ = basic_client.request("GET", "/admin/manage/add-company")
    check("basic gets the forbidden page", status == 403, status)

    # Languages.
    status, body = api(base, root, "/api/admin/manage/add-company", {"name": ""}, lang="pt-BR")
    check("api answers in Portuguese", status == 400 and body.get("error") == "Digite o nome da empresa.", body)
    status, body = api(base, root, "/api/admin/manage/add-company", {"name": ""})
    check("api answers in English by default", status == 400 and body.get("error") == "Enter the company's name.", body)
    status, _, _ = root_client.request("POST", "/admin/set-language", {"language": "pt", "next": "/admin/erp/dashboard"})
    status, page, _ = root_client.request("GET", "/admin/erp/dashboard")
    check("pages in Portuguese", status == 200 and "Painel" in page and 'lang="pt-BR"' in page, status)

    # The forms of the pages write what the api writes.
    def form(client, path, data):
        status, page, headers = client.request("POST", path, data)
        return status, headers.get("Location", ""), page
    status, location, _ = form(admin_client, "/admin/manage/add-transaction", {
        "type": "expense", "amount": "1.234,56", "description": "Form bill", "company": str(acme), "account": str(checking),
        "category": str(groceries), "date": today, "due-date": today, "next": "/admin/erp/list-transactions"})
    check("form: add transaction redirects with a notice", status == 303 and "notice=added" in location, (status, location))
    bill = ok(api(base, root, "/api/admin/erp/list-transactions", {"search": "Form bill"}))["transactions"][0]
    check("form: amount read the Portuguese way", bill["amount"] == -123456 and bill["status"] == "pending", bill)
    status, location, _ = form(admin_client, "/admin/manage/settle-transaction/%d" % bill["id"], {"settled-date": today, "next": "/admin/erp/list-transactions"})
    bill = ok(api(base, root, "/api/admin/erp/get-transaction", {"id": bill["id"]}))["transaction"]
    check("form: settle", status == 303 and bill["status"] == "settled" and bill["settled-date"] == today, bill)
    status, location, page = form(admin_client, "/admin/manage/add-transaction", {"type": "expense", "amount": "abc", "company": str(acme), "account": str(checking), "date": today})
    check("form: a bad amount answers the form again", status == 400 and "alert" in page, status)
    status, location, _ = form(admin_client, "/admin/manage/add-transfer", {"from-account": str(checking), "to-account": str(savings), "from-company": str(acme), "to-company": str(personal), "amount": "10", "date": today, "settled": "on"})
    check("form: inter-company transfer", status == 303, (status, location))
    second = listing["invoices"][1]
    status, location, _ = form(admin_client, "/admin/manage/pay-card-invoice/%d" % card, {"month": second["month"], "from-account": str(checking), "date": today, "settled": "on"})
    paid = ok(api(base, root, "/api/admin/erp/get-card-invoice", {"id": card, "month": second["month"]}))["invoice"]
    check("form: pay invoice", status == 303 and "notice=paid" in location and paid["remaining"] == 0, (location, paid))
    status, location, _ = form(admin_client, "/admin/manage/add-recurrence", {"type": "income", "amount": "100", "description": "Weekly sale", "company": str(acme), "account": str(checking), "frequency": "weekly", "interval": "1", "start-date": today, "status": "active"})
    check("form: add recurrence", status == 303 and "notice=added" in location, (status, location))
    status, location, _ = form(admin_client, "/admin/manage/add-budget", {"company": str(acme), "category": str(groceries), "period-start": today, "period-end": today, "amount": "50"})
    check("form: add budget", status == 303, (status, location))
    status, location, _ = form(admin_client, "/admin/manage/add-company", {"name": "Temporary", "status": "active"})
    temporary = next(c for c in ok(api(base, root, "/api/admin/erp/list-companies", {}))["companies"] if c["name"] == "Temporary")
    status, location, _ = form(admin_client, "/admin/manage/remove-company/%d" % temporary["id"], {})
    check("form: remove an unused company", status == 303 and "notice=removed" in location, (status, location))
    status, location, _ = form(admin_client, "/admin/manage/remove-company/%d" % acme, {})
    check("form: removing a used company is refused", status == 303 and "notice=in-use" in location, (status, location))
    status, location, _ = form(basic_client, "/admin/manage/add-company", {"name": "Nope"})
    check("form: basic cannot post", status == 403, status)
    status, location, _ = form(admin_client, "/admin/set-language", {"language": "pt", "next": "https://evil.example/"})
    check("language switch never leaves the site", status == 303 and location.endswith("/admin/erp/dashboard"), location)

    # Every page answers.
    pages = ["/admin/erp/dashboard", "/admin/erp/list-transactions", "/admin/erp/payables-receivables", "/admin/erp/list-credit-cards",
             "/admin/erp/list-card-invoices/%d" % card, "/admin/erp/card-invoice/%d" % card, "/admin/erp/list-budgets", "/admin/erp/list-recurrences",
             "/admin/erp/cash-flow", "/admin/erp/projection", "/admin/erp/company-report", "/admin/erp/account-report?account=%d" % checking,
             "/admin/erp/budget-report", "/admin/erp/list-companies", "/admin/erp/list-accounts", "/admin/erp/list-categories",
             "/admin/erp/list-third-parties", "/admin/manage/add-transaction", "/admin/manage/add-transfer", "/admin/manage/add-company",
             "/admin/manage/add-account", "/admin/manage/add-category", "/admin/manage/add-third-party", "/admin/manage/add-recurrence",
             "/admin/manage/add-budget", "/admin/manage/set-account/%d" % card, "/admin/root/settings", "/admin/home",
             "/admin/list-backoffice-users", "/admin/root/list-backups"]
    for client, lang in [(root_client, "pt"), (admin_client, "en")]:
        for path in pages:
            status, page, _ = client.request("GET", path)
            expected = 403 if client is admin_client and path.startswith("/admin/root/") else 200
            check("page %s (%s)" % (path, lang), status == expected and ("<html" in page), status)


if __name__ == "__main__":
    main()
