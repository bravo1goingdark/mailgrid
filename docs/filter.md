# 📤 `--filter` Flag – Advanced Recipient Filtering

Mailgrid supports advanced recipient filtering using logical expressions — enabling you to **target the right audience
directly from your CSV or Sheet**, without any preprocessing.

This feature uses a **mini expression engine** similar to SQL or logical formulas, allowing you to write complex conditions to precisely select recipients based on multiple criteria.

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
  --batch-size 5 \
  --filter 'tier = "pro" and subscribed = "true"'
```

---

## 🎨 Complex Filter Examples

### Multiple OR Conditions
```bash
# Target VIP customers or users from specific companies
--filter '(tier = "vip" or tier = "premium" or tier = "enterprise") and active = "true"'
```

### Combining Text and Numeric Filters
```bash
# High-value customers from tech companies
--filter 'score >= 85 and (company contains "tech" or company contains "software") and revenue > 10000'
```

### Complex Email Domain Filtering
```bash
# Exclude common personal email domains, focus on business emails
--filter 'not (email endswith "@gmail.com" or email endswith "@yahoo.com" or email endswith "@hotmail.com") and company != ..'
```

### Geographic and Demographic Targeting
```bash
# Target young professionals in specific cities
--filter '(location = "New York" or location = "San Francisco" or location = "London") and age >= 25 and age <= 40 and role contains "manager"'
```

### Engagement-Based Filtering
```bash
# Re-engagement campaign for inactive users with high potential
--filter 'last_login <= 30 and signup_score > 75 and tier != "basic" and email_opt_out != "true"'
```

### Empty Field Handling
```bash
# Only users with complete profiles
--filter 'name != .. and company != .. and phone != .. and location != ..'
```

---

## 💡 Best Practices & Tips

### ✅ Do's
- **Use parentheses** for complex expressions to ensure correct evaluation order
- **Quote string values** containing spaces or special characters: `name = "John Doe"`
- **Test filters** on small datasets first to verify your logic
- **Use `!= ..`** to ensure required fields are not empty
- **Combine conditions logically** – use `and` for restrictive filters, `or` for inclusive ones
- **Case doesn't matter** – all comparisons are case-insensitive

### ⚠️ Common Pitfalls

| Pitfall | Wrong | Right |
|---------|--------|--------|
| Missing quotes for multi-word values | `name = John Doe` | `name = "John Doe"` |
| Wrong operator precedence | `tier = "vip" or tier = "pro" and active = "true"` | `(tier = "vip" or tier = "pro") and active = "true"` |
| Confusing empty check | `company = ""` | `company = ..` |
| Using wrong numeric comparison | `score = "80"` | `score = 80` or `score >= 80` |
| Forgetting negation grouping | `not email contains gmail or yahoo` | `not (email contains "gmail" or email contains "yahoo")` |

### 📝 Field Name Guidelines
- Field names should match your CSV column headers exactly
- Spaces in field names work: `"Last Login" = ..`
- Special characters need quoting: `"email-verified" = "true"`
- Case-insensitive field matching: `EMAIL` = `email` = `Email`

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

---

## 🎨 Advanced Filtering Patterns

### Segmentation by Multiple Criteria
```bash
# Enterprise customers in North America who haven't been contacted recently
--filter '(tier = "enterprise" or revenue >= 100000) and (country = "US" or country = "Canada") and last_contact <= 30'
```

### A/B Testing Groups
```bash
# Split users into test groups based on user ID modulo
--filter 'user_id % 2 = 0'  # Group A (even IDs)
--filter 'user_id % 2 = 1'  # Group B (odd IDs)
```

### Exclude Previously Contacted
```bash
# New campaign excluding users from previous campaigns
--filter 'campaign_2024_q1 != "sent" and campaign_2024_q2 != "sent" and active = "true"'
```

### Lifecycle Stage Targeting
```bash
# Target users in specific lifecycle stages
--filter '(signup_days >= 7 and signup_days <= 30) or (last_purchase_days >= 60 and last_purchase_days <= 180)'
```

### Multi-Language Support
```bash
# Target users by preferred language
--filter 'language = "en" or language = "english" or preferred_locale contains "en"'
```

---

## 🎯 Real-World Use Cases

### 1. Product Launch Announcement
```bash
# Target engaged users who might be interested in new features
--filter 'engagement_score >= 70 and (tier = "pro" or tier = "premium") and last_login <= 7'
```

### 2. Win-Back Campaign
```bash
# Re-engage churned premium users
--filter 'tier = "premium" and last_login > 90 and signup_date <= 365 and email_opt_out != "true"'
```

### 3. Onboarding Follow-up
```bash
# New users who haven't completed setup
--filter 'signup_days >= 3 and signup_days <= 14 and onboarding_complete != "true" and trial_active = "true"'
```

### 4. Renewal Reminders
```bash
# Users with subscriptions expiring soon
--filter 'subscription_expires <= 30 and subscription_expires > 0 and auto_renew != "true"'
```

### 5. Event Invitations
```bash
# Local users in target industries for event
--filter '(city = "San Francisco" or city = "San Jose" or city = "Oakland") and (industry contains "tech" or industry contains "startup") and event_opt_in = "true"'
```

---

## 🛠️ Debugging Your Filters

### Use Dry-Run Mode
```bash
# Test your filter without sending emails
mailgrid --csv contacts.csv --filter 'your_filter_here' --dry-run
```

### Start Simple, Then Build
```bash
# Step 1: Test individual conditions
--filter 'tier = "pro"'                    # Should match pro users
--filter 'active = "true"'                 # Should match active users

# Step 2: Combine with AND
--filter 'tier = "pro" and active = "true"' # Should match active pro users

# Step 3: Add complexity
--filter '(tier = "pro" or tier = "premium") and active = "true" and last_login <= 30'
```

### Common Debug Techniques
- **Check field names**: Ensure they match your CSV headers exactly
- **Verify data types**: Use `score = 80` for numbers, `score = "80"` for text
- **Test edge cases**: Empty fields, special characters, very long strings
- **Use parentheses liberally**: Better safe than sorry with operator precedence

---

## ⚡ Performance Tips

### Filter Efficiency Guidelines

| ✅ **Efficient** | ❌ **Less Efficient** | **Why** |
|-------------|---------------------|----------|
| `tier = "pro"` | `name contains ""` | Exact matches are faster than substring searches |
| `active = "true"` | `not active = "false"` | Positive conditions process faster |
| `score >= 80 and tier = "pro"` | `tier = "pro" and score >= 80` | Put most selective conditions first |
| `company != ..` | `company contains "something"` | Empty checks are faster than pattern matching |

### Large Dataset Optimization
```bash
# For large CSVs (>10k records), consider:
# 1. Most selective condition first
--filter 'rare_field = "specific_value" and common_condition = "true"'

# 2. Avoid complex nested conditions if possible
--filter 'tier = "enterprise"' # Instead of: (tier = "pro" or tier = "premium" or tier = "enterprise")

# 3. Use numeric comparisons when possible
--filter 'score >= 85' # Instead of: score_category = "high"
```

### Memory Considerations
- Filters are applied during CSV processing, so memory usage scales with dataset size
- Complex filters with many OR conditions may increase processing time
- Consider preprocessing very large datasets if filters are too complex

---

## 💡 Quick Reference Cheat Sheet

```bash
# Text Matching
email contains "gmail"          # Substring match
name startswith "Dr."           # Prefix match  
domain endswith ".com"          # Suffix match
company = "Acme Corp"           # Exact match
status != "inactive"           # Not equal

# Numeric Comparisons
score > 80                     # Greater than
age >= 25                      # Greater than or equal
visits < 10                    # Less than
revenue <= 50000              # Less than or equal

# Empty Field Checks
company != ..                  # Field is not empty
phone = ..                     # Field is empty

# Logical Operators
condition1 and condition2      # Both must be true
condition1 or condition2       # At least one must be true
not condition                  # Invert condition
(group1) and (group2)          # Control precedence

# Real Examples
--filter 'tier = "premium" and last_login <= 30'
--filter '(company contains "tech" or industry = "software") and revenue > 100000'
--filter 'not (email endswith "@test.com" or email endswith "@example.com")'
```

---

> 📝 **Note:** Comparison support for **dates and timestamps** (e.g., `created_at > "2024-01-01"` or `sent_at <= "2025-07-27T15:00:00"`) is **not yet implemented**.
>
> This feature is planned for a future release of Mailgrid and will allow filtering using:
> - Standard date formats (`YYYY-MM-DD`)
> - Timestamps with time (`YYYY-MM-DDTHH:MM:SS`)
> - Operators like `>`, `<`, `>=`, `<=` for temporal filtering
