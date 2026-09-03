package syntax

import "bytes"

// Tree is deliberately lossless: tokens include trivia and error tokens, while
// Source retains the exact original bytes. Typed projections live outside it.
type Tree struct {
	Source []byte
	Tokens []Token
}

func (t Tree) Bytes() []byte {
	var out bytes.Buffer
	for _, token := range t.Tokens {
		if token.Kind != TokenEOF {
			out.WriteString(token.Text)
		}
	}
	return out.Bytes()
}
