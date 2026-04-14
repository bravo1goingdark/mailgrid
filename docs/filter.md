# Mailgrid — Filter Expression Reference

> Complete syntax reference, operator catalog, and use-case examples for `--filter`

---

## Table of Contents

- [Overview](#overview)
- [Comparison Operators](#comparison-operators)
  - [Equality (`==`)](#equality-)
  - [Inequality (`!=`)](#inequality-)
  - [Numeric Ordering (`>`, `>=`, `<`, `<=`)](#numeric-ordering----)
- [String Operators](#string-operators)
  - [`contains`](#contains)
  - [`startsWith`](#startswith)
  - [`endsWith`](#endswith)
- [Logical Operators](#logical-operators)
  - [`&&` / `and`](#--and)
  - [`||` / `or`](#--or)
  - [`!` / `not`](#--not)
- [Grouping](#grouping)
- [Truthy / Empty Checks](#truthy--empty-checks)
- [Examples by Use Case](#examples-by-use-case)
- [Testing Filters](#testing-filters)
- [Avoid / Prefer](#avoid--prefer)

---

## Overview

`--filter` (`-F`) accepts a logical expression evaluated against every recipient row **before** any emails are prepared or sent. Recipients that do not match are silently skipped; the final log line reports the skip count.

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium" && country == "US"'
```

**How fields are resolved:**

- Column names from the CSV header are normalized to lowercase. `FirstName`, `FIRSTNAME`, and `firstname` all resolve as `firstname`.
- String literals in the expression are also lowercased before comparison, so `tier == "Premium"` and `tier == "premium"` are identical.
- The `email` column is always available even if the CSV does not have a column named `email`.
- A field that does not exist in the CSV evaluates as an empty string (falsy). No error is raised — the recipient is skipped.

---

## Comparison Operators

### Equality (`==`)

```
field == "value"
```

**Behavior:** Returns `true` when the field's value, after lowercasing, equals the literal string.

**Example:**

```bash
--filter 'tier == "premium"'
--filter 'status == "active"'
--filter 'country == "US"'
```

---

### Inequality (`!=`)

```
field != "value"
```

**Behavior:** Returns `true` when the field's value does not equal the literal string. Absent fields evaluate as `""`, so `field != ""` is a non-empty check.

**Example:**

```bash
--filter 'status != "inactive"'
--filter 'plan != "trial"'
--filter 'company != ""'       # recipient has a company value
```

---

### Numeric Ordering (`>`, `>=`, `<`, `<=`)

```
field > number
field >= number
field < number
field <= number
```

**Behavior:** CSV values are strings; Mailgrid coerces them to float64 before the comparison. If the field value cannot be parsed as a number, the comparison returns `false` and the recipient is skipped.

**Arguments:**

| Operator | Meaning |
|---|---|
| `>` | strictly greater than |
| `>=` | greater than or equal |
| `<` | strictly less than |
| `<=` | less than or equal |

**Example:**

```bash
--filter 'age >= 18'
--filter 'score > 80'
--filter 'days_remaining > 0'
--filter 'score > 80 && score < 100'
```

**Notes:**
- String operators (`contains`, `startsWith`, `endsWith`) still work on numeric-looking columns — they match textually. `score contains "8"` matches `80`, `18`, and `800`. Use numeric operators when you need numeric semantics.

---

## String Operators

Two equivalent syntaxes are accepted: operator form and function form. Both behave identically.

### `contains`

```
field contains "substring"
contains(field, "substring")
```

**Behavior:** Returns `true` when the field value contains the substring. Case-insensitive.

**Example:**

```bash
--filter 'email contains "@gmail.com"'
--filter 'contains(company, "tech")'
--filter 'industry contains "software" || industry contains "saas"'
```

---

### `startsWith`

```
field startsWith "prefix"
startsWith(field, "prefix")
```

**Behavior:** Returns `true` when the field value begins with the prefix. Case-insensitive.

**Example:**

```bash
--filter 'name startsWith "Dr."'
--filter 'startsWith(email, "admin")'
--filter 'startsWith(role, "super")'
```

---

### `endsWith`

```
field endsWith "suffix"
endsWith(field, "suffix")
```

**Behavior:** Returns `true` when the field value ends with the suffix. Case-insensitive.

**Example:**

```bash
--filter 'email endsWith "@acme.com"'
--filter 'email endsWith "@corp.com" || email endsWith "@enterprise.io"'
```

---

## Logical Operators

### `&&` / `and`

```
expr && expr
expr and expr
```

**Behavior:** Both sub-expressions must be true. Short-circuits on the first `false`.

**Example:**

```bash
--filter 'tier == "premium" && country == "US"'
--filter 'tier == "premium" and country == "US"'
```

---

### `||` / `or`

```
expr || expr
expr or expr
```

**Behavior:** At least one sub-expression must be true. Short-circuits on the first `true`.

**Example:**

```bash
--filter 'tier == "vip" || tier == "premium"'
--filter 'tier == "vip" or tier == "premium"'
```

---

### `!` / `not`

```
!expr
not expr
```

**Behavior:** Inverts the boolean result of the sub-expression.

**Example:**

```bash
--filter '!email contains "@test.com"'
--filter 'not email contains "@test.com"'
--filter '!company'         # true when company is empty
```

**Precedence:**

`!` binds tighter than `&&`, which binds tighter than `||`. Use parentheses whenever the intent could be ambiguous.

---

## Grouping

```
(expr)
```

**Behavior:** Parentheses override default precedence and force the enclosed expression to be evaluated first.

**Example:**

```bash
# Without parentheses, && binds first: tier=="vip" OR (tier=="premium" AND country=="US")
--filter 'tier == "vip" || tier == "premium" && country == "US"'

# With parentheses: (tier=="vip" OR tier=="premium") AND country=="US"
--filter '(tier == "vip" || tier == "premium") && country == "US"'

--filter '(country == "US" || country == "CA") && age >= 18'
--filter 'status == "active" && (score > 90 || tier == "enterprise")'
```

**Notes:**
- Parentheses can be nested arbitrarily deep.
- Always add parentheses when mixing `&&` and `||` — the implicit precedence rules are easy to misread.

---

## Truthy / Empty Checks

A field name used as a standalone expression evaluates as `true` when the field is non-empty and `false` when it is empty or absent.

```bash
--filter 'company'          # passes when company column is non-empty
--filter '!company'         # passes when company column is empty or absent
```

This is equivalent to:

```bash
--filter 'company != ""'
--filter 'company == ""'
```

---

## Examples by Use Case

### E-commerce

```bash
# Recent purchasers in the US
--filter 'purchase_count > 0 && country == "US"'

# High-value customers who are active
--filter 'total_spent >= 1000 && status == "active"'

# Abandoned cart — has items, hasn't purchased recently
--filter 'cart_items > 0 && days_since_purchase > 30'

# Re-engagement: inactive for 60+ days
--filter 'days_since_last_login > 60 && status == "active"'
```

### SaaS

```bash
# Trial users about to expire
--filter 'plan == "trial" && days_left <= 3'

# Enterprise accounts only
--filter 'company_size > 100 && tier == "enterprise"'

# Users who haven't completed onboarding
--filter 'onboarding_complete == "false" && plan != "trial"'

# Churned users for win-back campaign
--filter 'status == "churned" && days_since_churn <= 90'
```

### Newsletter & Content

```bash
# Subscribed users excluding test addresses
--filter 'subscribed == "true" && !email contains "@test.com"'

# Premium newsletter subscribers
--filter 'newsletter == "premium" || tier == "vip"'

# Specific region
--filter 'region == "EU" && language == "en"'

# Tech industry contacts
--filter 'industry contains "tech" || industry contains "software"'
```

### Domain / Email Targeting

```bash
# Only Gmail users
--filter 'email contains "@gmail.com"'

# Corporate only — exclude free providers
--filter '!email contains "@gmail.com" && !email contains "@yahoo.com" && !email contains "@hotmail.com"'

# Specific company domain
--filter 'email endsWith "@acme.com"'

# Multiple allowed domains
--filter 'email endsWith "@acme.com" || email endsWith "@beta.com"'
```

### Role-Based Targeting

```bash
# Decision makers
--filter 'title contains "CEO" || title contains "CTO" || title contains "VP"'

# Admins and owners
--filter 'role == "admin" || role == "owner"'

# All except read-only members
--filter 'role != "viewer"'
```

### Geography

```bash
# North America only
--filter 'country == "US" || country == "CA"'

# Exclude EU for compliance
--filter 'region != "EU"'

# City-level targeting
--filter 'contains(city, "new york") || contains(city, "los angeles")'
```

### Combined / Complex

```bash
# Premium US users with recent activity, excluding test emails
--filter '(tier == "premium" || tier == "vip") && country == "US" && days_since_login <= 30 && !email contains "@test.com"'

# Trial users expiring soon with a company on file
--filter 'plan == "trial" && days_left <= 3 && company != ""'

# Referral campaign: referred users who haven't converted
--filter 'referral_source != "" && plan == "free" && signup_age_days > 7'
```

---

## Testing Filters

Always validate with `--dry-run` before a live send. No emails are sent; the output shows every matched recipient with their rendered subject and body.

```bash
mailgrid --env config.json --csv recipients.csv --template email.html \
  --filter 'tier == "premium" && country == "US"' \
  --dry-run
```

**Recommended workflow:**

```bash
# 1. Count matched recipients
mailgrid ... --filter 'your expression' --dry-run 2>&1 | grep "Email #" | wc -l

# 2. Inspect a rendered email for one matched recipient
mailgrid ... --filter 'your expression' --preview

# 3. Send
mailgrid ... --filter 'your expression'
```

**Diagnosing a zero match count:**

- Check for column name casing. `FirstName` in the CSV header becomes `firstname` in the filter. Run with `--log-level debug` to see the field map for each recipient.
- Check that string literals are double-quoted. `tier == premium` treats `premium` as a field reference, not a string, and evaluates as `false` for every recipient.
- An absent field evaluates as an empty string, not an error. A typo in a field name silently skips all rows.

---

## Avoid / Prefer

**Avoid: bare field names as strings**

```bash
# Wrong — 'premium' is treated as a field reference
--filter 'tier == premium'

# Correct — wrap string literals in double quotes
--filter 'tier == "premium"'
```

---

**Avoid: single-quoting inside the expression**

```bash
# Wrong — shell interprets the single quote inside
--filter 'name startsWith "O'Brien"'

# Correct — use $'...' syntax for expressions with single quotes
--filter $'name startsWith "O\'Brien"'
```

---

**Avoid: uppercase field names**

```bash
# Wrong — FirstName in the CSV becomes 'firstname' in the filter
--filter 'FirstName startsWith "Alice"'

# Correct
--filter 'firstname startsWith "Alice"'
```

---

**Avoid: using string operators for numeric comparisons**

```bash
# Wrong — 'score contains "8"' matches 80, 18, and 800
--filter 'score contains "8"'

# Correct — use numeric operators for numeric semantics
--filter 'score >= 80 && score < 90'
```

---

**Avoid: mixing && and || without parentheses**

```bash
# Ambiguous — && binds first, may not be what you want
--filter 'tier == "vip" || tier == "premium" && country == "US"'

# Explicit — intent is clear
--filter '(tier == "vip" || tier == "premium") && country == "US"'
```

---

**Prefer: shell single-quoting the whole expression**

```bash
# Wrong — shell expands && before mailgrid sees it
mailgrid ... --filter tier == "premium" && country == "US"

# Correct — single quotes prevent shell interpretation of &&, ||, !, >
mailgrid ... --filter 'tier == "premium" && country == "US"'
```
