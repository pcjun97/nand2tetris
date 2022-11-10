package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type TokenType int
type KeyWord int

const (
	TOKEN_TYPE_ERROR TokenType = iota
	KEYWORD
	SYMBOL
	IDENTIFIER
	INT_CONST
	STRING_CONST
)

const (
	KEYWORD_ERROR KeyWord = iota
	CLASS
	METHOD
	FUNCTION
	CONSTRUCTOR
	INT
	BOOLEAN
	CHAR
	VOID
	VAR
	STATIC
	FIELD
	LET
	DO
	IF
	ELSE
	WHILE
	RETURN
	TRUE
	FALSE
	NULL
	THIS
)

var keywordMap map[string]KeyWord = map[string]KeyWord{
	"class":       CLASS,
	"method":      METHOD,
	"function":    FUNCTION,
	"constructor": CONSTRUCTOR,
	"int":         INT,
	"boolean":     BOOLEAN,
	"char":        CHAR,
	"void":        VOID,
	"var":         VAR,
	"static":      STATIC,
	"field":       FIELD,
	"let":         LET,
	"do":          DO,
	"if":          IF,
	"else":        ELSE,
	"while":       WHILE,
	"return":      RETURN,
	"true":        TRUE,
	"false":       FALSE,
	"null":        NULL,
	"this":        THIS,
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
		tokenType: TOKEN_TYPE_ERROR,
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
					log.Fatal(err)
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
	if t.tokenType == KEYWORD {
		return keywordMap[t.current]
	}

	return KEYWORD_ERROR
}

func (t *Tokenizer) Symbol() rune {
	if t.tokenType == SYMBOL {
		return rune(t.current[0])
	}

	return -1
}

func (t *Tokenizer) Identifier() string {
	if t.tokenType == IDENTIFIER {
		return t.current
	}

	return ""
}

func (t *Tokenizer) IntVal() int {
	if t.tokenType == INT_CONST {
		val, err := strconv.ParseInt(t.current, 10, 16)
		if err != nil {
			log.Fatalln(err)
		}

		return int(val)
	}

	return 0
}

func (t *Tokenizer) StringVal() string {
	if t.tokenType == STRING_CONST {
		return t.current[1 : len(t.current)-1]
	}

	return ""
}

func (t *Tokenizer) setTokenType() {
	if t.current[0] == '"' {
		t.tokenType = STRING_CONST
		return
	}

	if strings.ContainsRune("0123456789", rune(t.current[0])) {
		t.tokenType = INT_CONST
		return
	}

	if len(t.current) == 1 && strings.ContainsRune("{}()[].,;+-*/&|<>=~", rune(t.current[0])) {
		t.tokenType = SYMBOL
		return
	}

	if _, ok := keywordMap[t.current]; ok {
		t.tokenType = KEYWORD
		return
	}

	if !strings.ContainsAny(t.current, "{}()[].,;+-*/&|<>=~") {
		t.tokenType = IDENTIFIER
		return
	}

	t.tokenType = TOKEN_TYPE_ERROR
}
