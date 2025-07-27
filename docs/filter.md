# 📤 `--filter` Flag – Advanced Recipient Filtering

Mailgrid supports advanced recipient filtering using logical expressions — enabling you to **target the right audience
directly from your CSV or Sheet**, without any preprocessing.

This feature uses a **mini expression engine** similar to SQL or logical formulas.

---

## ✅ Basic Usage

Use the `--filter` flag to apply a logical expression: \
All are case-insensitive

```bash
mailgrid \
  --csv contacts.csv \
  --env config.json \
  --template welcome.html \
  --subject "Hi {{ .name }}" \
  --concurrency 5 \
  --batch-size 5
  --filter 'tier = "pro" and subscribed = "true"'
```

---

## 🔧 Supported Filter Operators in Mailgrid

| Operator     | Description                       | Example                                         |
|--------------|-----------------------------------|-------------------------------------------------|
| `=` / `==`   | Field equals value                | `company = "Gadgetry"`                          |
| `!=`         | Field not equal to value          | `tier != "basic"`                               |
| `contains`   | Field contains substring          | `name contains "ash"`                           |
| `startswith` | Field starts with substring       | `email startswith "kumar"`                      |
| `endswith`   | Field ends with substring         | `email endswith "@gmail.com"`                   |
| `>` / `>=`   | Greater than (numeric comparison) | `score > 80`                                    |
| `<` / `<=`   | Less than (numeric comparison)    | `age <= 30`                                     |
| `..`         | Empty value                       | `company = ..`                                  |
| `not`        | Negation of a condition           | `not email endswith "@example.com"`             |
| `and`, `or`  | Logical AND / OR combinations     | `tier = "pro" and age > 25`                     |
| `()`         | Group expressions for precedence  | `(score > 50 and tier = "pro") or company = ..` |

---

## 🔁 Logical Operators

| Operator | Alias                     | Description                             |
|----------|---------------------------|-----------------------------------------|
| `and`    | `&&`                      | Both conditions must be true            |
| `or`     | <code>&#124;&#124;</code> | At least one condition must be true     |
| `not`    | `!`                       | Inverts the result of the condition     |
| `()`     | —                         | Groups conditions to control precedence |

---

### ✅ `and` / `&&` — Logical AND

Use this to match only when **both conditions are true**.

#### Example:

```bash
--filter 'company != mailgrid and email endswith "@gmail.com"'
````

---

### 🔁 `or` / `||` — Logical OR

Use this to match when **at least one condition is true**.

#### Example:

```bash
--filter 'tier = "vip" or tier = "premium"'
```

--- 

### 🚫 `not` / `!` — Logical NOT

Use this to **invert** a condition — it matches when the condition is **false**.

#### Example:

```bash
--filter 'not email endswith "@example.com"'
```

### 🧠 `()` — Grouped Expressions

Use parentheses to **group conditions** and control the **order of evaluation**.

#### Example:

```bash
--filter '(tier = "vip" or tier = "premium") and location = "India"'

```

---

## 🧮 Comparison Operators

| Operator     | Alias | Description                                        | Example                         |
|--------------|-------|----------------------------------------------------|---------------------------------|
| `=`          | `==`  | Checks if the field **exactly equals** a value     | `name = "Aakash"`               |
| `!=`         | —     | Checks if the field **does not equal** a value     | `company != "Google"`           |
| `contains`   | —     | Case-insensitive match if field **contains** value | `email contains "gmail"`        |
| `startswith` | —     | Checks if field **starts with** a value            | `name startswith "Aa"`          |
| `endswith`   | —     | Checks if field **ends with** a value              | `email endswith "@example.com"` |
| `>`          | —     | Checks if field is **greater than** value          | `score > 80`                    |
| `<`          | —     | Checks if field is **less than** value             | `age < 60`                      |
| `>=`         | —     | Checks if field is **greater than or equal**       | `salary >= 50000`               |
| `<=`         | —     | Checks if field is **less than or equal**          | `visits <= 10`                  |
| `!= ..`      | —     | Checks if field is **non-empty**                   | `company != ..`                 |

---

### 🎯 `=` / `==` — Equals

Use this to match when a field **exactly equals** the given value (case-insensitive).

#### Example:

```bash
--filter 'name = Aakash'
```

---

### ❌ `!=` — Not Equals

Use this to match when a field **does not equal** the given value (case-insensitive).

#### Example:

```bash
--filter 'company != Google'

```
---
### 🔍 `contains` — Substring Match

Use this to match when a field **contains** the given substring (case-insensitive).

#### Example:

```bash
--filter 'email contains gmail.com'
```

---

### 🔼 `startswith` — Prefix Match

Use this to match when a field **starts with** the given value (case-insensitive).

#### Example:

```bash
--filter 'name startswith "Dr."'
```
---
### 🔽 `endswith` — Suffix Match

Use this to match when a field **ends with** the given value (case-insensitive).

#### Example:

```bash
--filter 'email endswith "@example.com"'
```

---

### 🔼 `>` / `>=` — Greater Than / Greater Than or Equal

Use these to match when a field is **numerically greater** than (or equal to) a given value.

#### Examples:

```bash
--filter 'score > 80'
```
or

```bash
--filter 'salary >= 50000'
```

---
### 🔽 `<` / `<=` — Less Than / Less Than or Equal

Use these to match when a field is **numerically less** than (or equal to) a given value.

#### Examples:
```bash
--filter 'age < 60'
```

or

```bash
--filter 'visits <= 10'

```
---
### 📭 `!= ..` — Non-Empty Field

Use this to match rows where a field is **not empty** or missing.

#### Example:
```bash
--filter 'company != ..'
```

---

> 📝 **Note:** Comparison support for **dates and timestamps** (e.g., `created_at > "2024-01-01"` or `sent_at <= "2025-07-27T15:00:00"`) is **not yet implemented**.
>
> This feature is planned for a future release of Mailgrid and will allow filtering using:
> - Standard date formats (`YYYY-MM-DD`)
> - Timestamps with time (`YYYY-MM-DDTHH:MM:SS`)
> - Operators like `>`, `<`, `>=`, `<=` for temporal filtering
