# Filter Documentation

Filter recipients using logical expressions to target specific audiences.

## Basic Usage

```bash
mailgrid --csv recipients.csv --template email.html --filter 'tier == "premium"'
```

All comparisons are **case-insensitive**.

## Operators

### Comparison Operators

| Operator | Alias | Description | Example |
|----------|-------|-------------|---------|
| `==` | `=` | Equals | `tier == "premium"` |
| `!=` | | Not equals | `status != "inactive"` |
| `contains` | | Contains substring | `email contains "@gmail"` |
| `startsWith` | `startswith` | Starts with | `name startsWith "Dr."` |
| `endsWith` | `endswith` | Ends with | `email endsWith "@corp.com"` |
| `>` | | Greater than | `score > 80` |
| `>=` | | Greater or equal | `age >= 18` |
| `<` | | Less than | `visits < 10` |
| `<=` | | Less or equal | `balance <= 0` |

### Logical Operators

| Operator | Alias | Description |
|----------|-------|-------------|
| `and` | `&&` | Both conditions must be true |
| `or` | `\|\|` | At least one must be true |
| `not` | `!` | Inverts the condition |
| `()` | | Groups conditions |

### Special Values

| Value | Description |
|--------|-------------|
| `..` | Empty/null value |
| `!..` | Non-empty value |

## Examples

### Simple Filters

```bash
# Premium users only
--filter 'tier == "premium"'

# Exclude inactive users
--filter 'status != "inactive"'

# Gmail users
--filter 'email contains "@gmail.com"'

# Specific domain
--filter 'email endsWith "@company.com"'
```

### Combined Conditions

```bash
# AND - both must match
--filter 'tier == "premium" and age > 25'

# OR - at least one matches
--filter 'tier == "vip" or tier == "premium"'

# NOT - exclude matches
--filter 'not email contains "@test.com"'
```

### Grouped Expressions

```bash
# Complex logic with parentheses
--filter '(tier == "vip" or tier == "premium") and location == "US"'

# Multiple AND/OR
--filter '(country == "US" or country == "CA") and age >= 18 and subscribed == true'
```

### Numeric Comparisons

```bash
# Age filter
--filter 'age >= 18'

# Score threshold
--filter 'score > 80'

# Balance check
--filter 'balance > 0'
```

### Empty Value Checks

```bash
# Has company field
--filter 'company != ..'

# No company field
--filter 'company == ..'
```

## Real-World Examples

### E-commerce

```bash
# Recent purchasers in US
--filter 'purchase_count > 0 and country == "US"'

# High-value customers
--filter 'total_spent >= 1000 and status == "active"'
```

### SaaS

```bash
# Trial users about to expire
--filter 'plan == "trial" and days_left <= 3'

# Enterprise accounts
--filter 'company_size > 100 and tier == "enterprise"'
```

### Newsletter

```bash
# Subscribed users excluding unsubscribed
--filter 'subscribed == true and not email contains "@test.com"'

# Premium newsletter subscribers
--filter 'newsletter == "premium" or tier == "vip"'
```

## Tips

- Use `--dry-run` with `--filter` to test your filter before sending
- Filters are evaluated case-insensitively
- Group complex expressions with parentheses for clarity
- Combine with `--concurrency` for faster processing of filtered lists
