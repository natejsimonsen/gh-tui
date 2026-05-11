package ui

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/lipgloss"
)

func getLexer(filename string) chroma.Lexer {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Match("file" + filepath.Ext(filename))
	}
	if lexer == nil {
		return nil
	}
	return chroma.Coalesce(lexer)
}

func highlightLine(lexer chroma.Lexer, text string, bg lipgloss.Color) string {
	if lexer == nil {
		return renderPlainWithBg(text, bg)
	}

	iter, err := lexer.Tokenise(nil, text)
	if err != nil {
		return renderPlainWithBg(text, bg)
	}

	var b strings.Builder
	for _, tok := range iter.Tokens() {
		if tok.Value == "\n" {
			continue
		}
		renderToken(&b, tok, bg)
	}
	return b.String()
}

func renderTokens(tokens []chroma.Token, bg lipgloss.Color) string {
	var b strings.Builder
	for _, tok := range tokens {
		renderToken(&b, tok, bg)
	}
	return b.String()
}

func renderToken(b *strings.Builder, tok chroma.Token, bg lipgloss.Color) {
	fg := tokenFg(tok.Type)
	if fg == "" {
		if bg != "" {
			b.WriteString(lipgloss.NewStyle().Background(bg).Render(tok.Value))
		} else {
			b.WriteString(tok.Value)
		}
	} else {
		s := lipgloss.NewStyle().Foreground(fg)
		if bg != "" {
			s = s.Background(bg)
		}
		b.WriteString(s.Render(tok.Value))
	}
}

func renderPlainWithBg(text string, bg lipgloss.Color) string {
	if bg != "" {
		return lipgloss.NewStyle().Background(bg).Render(text)
	}
	return text
}

func tokenizeLines(lexer chroma.Lexer, texts []string) [][]chroma.Token {
	if lexer == nil || len(texts) == 0 {
		return nil
	}
	full := strings.Join(texts, "\n")
	iter, err := lexer.Tokenise(nil, full)
	if err != nil {
		return nil
	}

	result := make([][]chroma.Token, len(texts))
	idx := 0
	for _, tok := range iter.Tokens() {
		remaining := tok.Value
		for remaining != "" && idx < len(result) {
			nl := strings.IndexByte(remaining, '\n')
			if nl < 0 {
				result[idx] = append(result[idx], chroma.Token{Type: tok.Type, Value: remaining})
				remaining = ""
			} else {
				if nl > 0 {
					result[idx] = append(result[idx], chroma.Token{Type: tok.Type, Value: remaining[:nl]})
				}
				idx++
				remaining = remaining[nl+1:]
			}
		}
	}
	return result
}

func tokenFg(tt chroma.TokenType) lipgloss.Color {
	t := activeTheme
	switch {
	case tt == chroma.Keyword || tt == chroma.KeywordDeclaration ||
		tt == chroma.KeywordNamespace || tt == chroma.KeywordReserved ||
		tt == chroma.KeywordPseudo:
		return t.Purple
	case tt == chroma.KeywordType || tt == chroma.KeywordConstant:
		return t.Cyan
	case tt == chroma.NameFunction || tt == chroma.NameFunctionMagic ||
		tt == chroma.NameDecorator:
		return t.Blue
	case tt == chroma.NameClass || tt == chroma.NameBuiltin ||
		tt == chroma.NameBuiltinPseudo || tt == chroma.NameException:
		return t.Cyan
	case tt == chroma.NameVariable || tt == chroma.NameVariableGlobal ||
		tt == chroma.NameVariableInstance || tt == chroma.NameVariableClass ||
		tt == chroma.NameVariableMagic:
		return t.Red
	case tt == chroma.NameConstant || tt == chroma.NameLabel:
		return t.Orange
	case tt == chroma.NameTag:
		return t.Red
	case tt == chroma.NameAttribute || tt == chroma.NameProperty:
		return t.Yellow
	case tt == chroma.NameNamespace || tt == chroma.NameEntity:
		return t.Purple
	case tt == chroma.NameOther:
		return t.Blue
	case tt == chroma.LiteralString || tt.InSubCategory(chroma.LiteralString):
		return t.Green
	case tt == chroma.LiteralNumber || tt.InSubCategory(chroma.LiteralNumber):
		return t.Orange
	case tt == chroma.Literal:
		return t.Green
	case tt == chroma.Comment || tt.InSubCategory(chroma.Comment):
		return t.FgMuted
	case tt == chroma.CommentPreproc:
		return t.Cyan
	case tt == chroma.Operator || tt == chroma.OperatorWord:
		return t.Red
	case tt == chroma.Punctuation:
		return t.FgSubtle
	case tt == chroma.GenericInserted:
		return t.Green
	case tt == chroma.GenericDeleted:
		return t.Red
	case tt == chroma.GenericHeading || tt == chroma.GenericSubheading:
		return t.Blue
	case tt == chroma.GenericEmph:
		return t.Yellow
	case tt == chroma.GenericStrong:
		return t.Orange
	case tt == chroma.GenericPrompt:
		return t.Cyan
	}
	return ""
}
