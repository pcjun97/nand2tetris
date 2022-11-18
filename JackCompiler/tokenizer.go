package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type TokenType int
type KeyWord int

const (
	TOKEN_TYPE_UNKNOWN TokenType = iota
	TOKEN_TYPE_KEYWORD
	TOKEN_TYPE_SYMBOL
	TOKEN_TYPE_IDENTIFIER
	TOKEN_TYPE_INT_CONST
	TOKEN_TYPE_STRING_CONST
)

var tokenTypeNames = []string{
	"TOKEN_TYPE_UNKNOWN",
	"TOKEN_TYPE_KEYWORD",
	"TOKEN_TYPE_SYMBOL",
	"TOKEN_TYPE_IDENTIFIER",
	"TOKEN_TYPE_INT_CONST",
	"TOKEN_TYPE_STRING_CONST",
}

func (t TokenType) String() string {
	return tokenTypeNames[t]
}

const (
	KEYWORD_UNKNOWN KeyWord = iota
	KEYWORD_CLASS
	KEYWORD_METHOD
	KEYWORD_FUNCTION
	KEYWORD_CONSTRUCTOR
	KEYWORD_INT
	KEYWORD_BOOLEAN
	KEYWORD_CHAR
	KEYWORD_VOID
	KEYWORD_VAR
	KEYWORD_STATIC
	KEYWORD_FIELD
	KEYWORD_LET
	KEYWORD_DO
	KEYWORD_IF
	KEYWORD_ELSE
	KEYWORD_WHILE
	KEYWORD_RETURN
	KEYWORD_TRUE
	KEYWORD_FALSE
	KEYWORD_NULL
	KEYWORD_THIS
)

var keywordNames = []string{
	"",
	"class",
	"method",
	"function",
	"constructor",
	"int",
	"boolean",
	"char",
	"void",
	"var",
	"static",
	"field",
	"let",
	"do",
	"if",
	"else",
	"while",
	"return",
	"true",
	"false",
	"null",
	"this",
}

func (k KeyWord) String() string {
	return keywordNames[k]
}

var stringToKeyword map[string]KeyWord = map[string]KeyWord{
	KEYWORD_CLASS.String():       KEYWORD_CLASS,
	KEYWORD_METHOD.String():      KEYWORD_METHOD,
	KEYWORD_FUNCTION.String():    KEYWORD_FUNCTION,
	KEYWORD_CONSTRUCTOR.String(): KEYWORD_CONSTRUCTOR,
	KEYWORD_INT.String():         KEYWORD_INT,
	KEYWORD_BOOLEAN.String():     KEYWORD_BOOLEAN,
	KEYWORD_CHAR.String():        KEYWORD_CHAR,
	KEYWORD_VOID.String():        KEYWORD_VOID,
	KEYWORD_VAR.String():         KEYWORD_VAR,
	KEYWORD_STATIC.String():      KEYWORD_STATIC,
	KEYWORD_FIELD.String():       KEYWORD_FIELD,
	KEYWORD_LET.String():         KEYWORD_LET,
	KEYWORD_DO.String():          KEYWORD_DO,
	KEYWORD_IF.String():          KEYWORD_IF,
	KEYWORD_ELSE.String():        KEYWORD_ELSE,
	KEYWORD_WHILE.String():       KEYWORD_WHILE,
	KEYWORD_RETURN.String():      KEYWORD_RETURN,
	KEYWORD_TRUE.String():        KEYWORD_TRUE,
	KEYWORD_FALSE.String():       KEYWORD_FALSE,
	KEYWORD_NULL.String():        KEYWORD_NULL,
	KEYWORD_THIS.String():        KEYWORD_THIS,
}

type Tokenizer struct {
	scanner   *bufio.Scanner
	tokenType TokenType
	next      string
	current   string
	line      string
	inComment bool
}

func NewTokenizer(input *os.File) *Tokenizer {
	scanner := bufio.NewScanner(input)

	t := Tokenizer{
		scanner:   scanner,
		tokenType: TOKEN_TYPE_UNKNOWN,
		next:      "init",
		current:   "",
		line:      "",
		inComment: false,
	}

	t.Advance()
	t.current = ""

	return &t
}

func (t *Tokenizer) HasMoreTokens() bool {
	return len(t.next) > 0
}

func (t *Tokenizer) Advance() {
	if t.HasMoreTokens() {
		t.current = t.next
		t.setTokenType()

		for {
			if len(t.line) > 0 {
				end := strings.IndexAny(t.line, " {}()[].,;+-*/&|<>=~\"")

				if end > 0 {
					t.next = t.line[0:end]
					t.line = t.line[end:]
				} else if end < 0 {
					t.next = t.line
					t.line = ""
				} else if t.line[0] == '"' {
					end = strings.IndexRune(t.line[1:], '"') + 1
					t.next = t.line[0 : end+1]
					t.line = t.line[end+1:]
				} else {
					t.next = t.line[0:1]
					t.line = t.line[1:]
				}

				t.line = strings.TrimSpace(t.line)
				break
			}

			if ok := t.scanner.Scan(); !ok {
				err := t.scanner.Err()
				if err != nil {
					panic(err)
				}

				t.next = ""
				break
			}

			line := t.scanner.Text()

			if t.inComment {
				blockCommentEnd := strings.Index(line, "*/")
				if blockCommentEnd > 0 {
					line = line[blockCommentEnd+2:]
					t.inComment = false
				} else {
					line = ""
				}
			}

			if !t.inComment {
				if comment := strings.Index(line, "//"); comment >= 0 {
					line = line[:comment]
				}

				for {
					blockCommentStart := strings.Index(line, "/*")
					if blockCommentStart < 0 {
						break
					}

					blockCommentEnd := strings.Index(line, "*/")
					if blockCommentEnd < 0 {
						t.inComment = true
						line = line[:blockCommentStart]
						break
					}

					line = line[:blockCommentStart] + " " + line[blockCommentEnd+2:]
				}
			}

			t.line = strings.TrimSpace(line)
		}
	}
}

func (t *Tokenizer) TokenType() TokenType {
	return t.tokenType
}

func (t *Tokenizer) KeyWord() KeyWord {
	if t.tokenType != TOKEN_TYPE_KEYWORD {
		panic("token type is not keyword: " + t.current)
	}

	return stringToKeyword[t.current]
}

func (t *Tokenizer) Symbol() rune {
	if t.tokenType != TOKEN_TYPE_SYMBOL {
		panic("token type is not symbol: " + t.current)
	}

	return rune(t.current[0])
}

func (t *Tokenizer) Identifier() string {
	if t.tokenType != TOKEN_TYPE_IDENTIFIER {
		panic("token type is not identifier: " + t.current)
	}

	return t.current
}

func (t *Tokenizer) IntVal() int {
	if t.tokenType != TOKEN_TYPE_INT_CONST {
		panic("token type is not intval: " + t.current)
	}

	val, err := strconv.ParseInt(t.current, 10, 16)
	if err != nil {
		panic(err)
	}

	return int(val)
}

func (t *Tokenizer) StringVal() string {
	if t.tokenType != TOKEN_TYPE_STRING_CONST {
		panic("token type is not stringval: " + t.current)
	}

	return t.current[1 : len(t.current)-1]
}

func (t *Tokenizer) setTokenType() {
	switch {
	case t.current[0] == '"':
		t.tokenType = TOKEN_TYPE_STRING_CONST

	case strings.ContainsRune("0123456789", rune(t.current[0])):
		t.tokenType = TOKEN_TYPE_INT_CONST

	case len(t.current) == 1 && strings.ContainsRune("{}()[].,;+-*/&|<>=~", rune(t.current[0])):
		t.tokenType = TOKEN_TYPE_SYMBOL

	case stringToKeyword[t.current] != 0 && stringToKeyword[t.current] != KEYWORD_UNKNOWN:
		t.tokenType = TOKEN_TYPE_KEYWORD

	case !strings.ContainsAny(t.current, "{}()[].,;+-*/&|<>=~"):
		t.tokenType = TOKEN_TYPE_IDENTIFIER

	default:
		panic("invalid token: " + t.current)
	}
}
