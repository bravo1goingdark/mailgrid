package expression

import (
	"fmt"
	"regexp"
	"strings"
)

// tokenRegex captures operators, parentheses, quoted strings, and non-whitespace tokens
var tokenRegex = regexp.MustCompile(`(?i)(\(|\)|!=|==|>=|<=|>|<|=|&&|\|\||!|contains|startswith|endswith|and|or|not|\"[^\"]*\"|[a-zA-Z0-9_.@+-]+)`)

// tokenize splits the input expression into tokens, unquoting quoted strings.
func tokenize(input string) []string {
	matches := tokenRegex.FindAllString(input, -1)
	tokens := make([]string, 0, len(matches))

	for _, tok := range matches {
		tok = strings.TrimSpace(tok)
		// Remove enclosing quotes from strings like "value"
		if len(tok) >= 2 && tok[0] == '"' && tok[len(tok)-1] == '"' {
			tok = tok[1 : len(tok)-1]
		}
		tokens = append(tokens, tok)
	}
	fmt.Printf("Tokens: %#v\n", tokens)

	return tokens
}
