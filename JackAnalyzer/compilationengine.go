package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type CompilationEngine struct {
	tokenizer *Tokenizer
	writer    *bufio.Writer
	indent    string
}

func NewCompilationEngine(input, output *os.File) *CompilationEngine {
	tokenizer := NewTokenizer(input)
	tokenizer.Advance()

	writer := bufio.NewWriter(output)

	c := CompilationEngine{
		tokenizer: tokenizer,
		writer:    writer,
		indent:    "",
	}

	return &c
}

func (c *CompilationEngine) CompileClass() {
	c.write("<class>\n")
	c.indent = c.indent + "  "

	c.processKeyword("class")
	c.processIdentifier()
	c.processSymbol('{')

	for c.tokenizer.TokenType() == KEYWORD {
		switch c.tokenizer.KeyWord() {
		case STATIC, FIELD:
			c.CompileClassVarDec()
		case CONSTRUCTOR, FUNCTION, METHOD:
			c.CompileSubroutine()
		}
	}

	c.processSymbol('}')

	c.indent = c.indent[2:]
	c.write("</class>\n")
}

func (c *CompilationEngine) CompileClassVarDec() {
	c.write("<classVarDec>\n")
	c.indent = c.indent + "  "

	switch c.tokenizer.KeyWord() {
	case STATIC:
		c.processKeyword("static")
	case FIELD:
		c.processKeyword("field")
	default:
		log.Fatalf("Syntax error: CompileClassVarDec: %s", c.tokenizer.current)
	}

	c.compileType()

	for {
		c.processIdentifier()
		if c.tokenizer.Symbol() == ';' {
			c.processSymbol(';')
			break
		}
		c.processSymbol(',')
	}

	c.indent = c.indent[2:]
	c.write("</classVarDec>\n")
}

func (c *CompilationEngine) CompileSubroutine() {
	c.write("<subroutineDec>\n")
	c.indent = c.indent + "  "

	switch c.tokenizer.KeyWord() {
	case CONSTRUCTOR:
		c.processKeyword("constructor")
	case FUNCTION:
		c.processKeyword("function")
	case METHOD:
		c.processKeyword("method")
	default:
		log.Fatalf("Syntax error: CompileSubroutine: %s", c.tokenizer.current)
	}

	switch {
	case c.tokenizer.TokenType() == KEYWORD && c.tokenizer.KeyWord() == VOID:
		c.processKeyword("void")
	default:
		c.compileType()
	}

	c.processIdentifier()
	c.processSymbol('(')
	c.CompileParameterList()
	c.processSymbol(')')
	c.CompileSubroutineBody()

	c.indent = c.indent[2:]
	c.write("</subroutineDec>\n")
}

func (c *CompilationEngine) CompileParameterList() {
	c.write("<parameterList>\n")
	c.indent = c.indent + "  "

	for c.tokenizer.TokenType() != SYMBOL {
		c.compileType()
		c.processIdentifier()

		if c.tokenizer.Symbol() == ',' {
			c.processSymbol(',')
		}
	}

	c.indent = c.indent[2:]
	c.write("</parameterList>\n")
}

func (c *CompilationEngine) CompileSubroutineBody() {
	c.write("<subroutineBody>\n")
	c.indent = c.indent + "  "

	c.processSymbol('{')

	for c.tokenizer.KeyWord() == VAR {
		c.CompileVarDec()
	}

	c.CompileStatements()
	c.processSymbol('}')

	c.indent = c.indent[2:]
	c.write("</subroutineBody>\n")
}

func (c *CompilationEngine) CompileVarDec() {
	c.write("<varDec>\n")
	c.indent = c.indent + "  "

	c.processKeyword("var")
	c.compileType()

	for {
		c.processIdentifier()
		if c.tokenizer.Symbol() == ';' {
			break
		}
		c.processSymbol(',')
	}

	c.processSymbol(';')

	c.indent = c.indent[2:]
	c.write("</varDec>\n")
}

func (c *CompilationEngine) CompileStatements() {
	c.write("<statements>\n")
	c.indent = c.indent + "  "

	for c.tokenizer.TokenType() == KEYWORD {
		switch c.tokenizer.KeyWord() {
		case LET:
			c.CompileLet()
		case IF:
			c.CompileIf()
		case WHILE:
			c.CompileWhile()
		case DO:
			c.CompileDo()
		case RETURN:
			c.CompileReturn()
		}
	}

	c.indent = c.indent[2:]
	c.write("</statements>\n")
}

func (c *CompilationEngine) CompileLet() {
	c.write("<letStatement>\n")
	c.indent = c.indent + "  "

	c.processKeyword("let")
	c.processIdentifier()

	if c.tokenizer.Symbol() == '[' {
		c.processSymbol('[')
		c.CompileExpression()
		c.processSymbol(']')
	}

	c.processSymbol('=')
	c.CompileExpression()
	c.processSymbol(';')

	c.indent = c.indent[2:]
	c.write("</letStatement>\n")
}

func (c *CompilationEngine) CompileIf() {
	c.write("<ifStatement>\n")
	c.indent = c.indent + "  "

	c.processKeyword("if")
	c.processSymbol('(')
	c.CompileExpression()
	c.processSymbol(')')
	c.processSymbol('{')
	c.CompileStatements()
	c.processSymbol('}')

	if c.tokenizer.KeyWord() == ELSE {
		c.processKeyword("else")
		c.processSymbol('{')
		c.CompileStatements()
		c.processSymbol('}')
	}

	c.indent = c.indent[2:]
	c.write("</ifStatement>\n")
}

func (c *CompilationEngine) CompileWhile() {
	c.write("<whileStatement>\n")
	c.indent = c.indent + "  "

	c.processKeyword("while")
	c.processSymbol('(')
	c.CompileExpression()
	c.processSymbol(')')
	c.processSymbol('{')
	c.CompileStatements()
	c.processSymbol('}')

	c.indent = c.indent[2:]
	c.write("</whileStatement>\n")
}

func (c *CompilationEngine) CompileDo() {
	c.write("<doStatement>\n")
	c.indent = c.indent + "  "

	c.processKeyword("do")
	c.processIdentifier()

	if c.tokenizer.Symbol() == '.' {
		c.processSymbol('.')
		c.processIdentifier()
	}

	c.processSymbol('(')
	c.CompileExpressionList()
	c.processSymbol(')')
	c.processSymbol(';')

	c.indent = c.indent[2:]
	c.write("</doStatement>\n")
}

func (c *CompilationEngine) CompileReturn() {
	c.write("<returnStatement>\n")
	c.indent = c.indent + "  "

	c.processKeyword("return")

	if c.tokenizer.TokenType() != SYMBOL || c.tokenizer.Symbol() != ';' {
		c.CompileExpression()
	}

	c.processSymbol(';')

	c.indent = c.indent[2:]
	c.write("</returnStatement>\n")
}

func (c *CompilationEngine) CompileExpression() {
	c.write("<expression>\n")
	c.indent = c.indent + "  "

	c.CompileTerm()

	for strings.ContainsRune("+-*/&|<>=", c.tokenizer.Symbol()) {
		c.processSymbol(c.tokenizer.Symbol())
		c.CompileTerm()
	}

	c.indent = c.indent[2:]
	c.write("</expression>\n")
}

func (c *CompilationEngine) CompileTerm() {
	c.write("<term>\n")
	c.indent = c.indent + "  "

	switch c.tokenizer.TokenType() {
	case INT_CONST:
		c.processIntConst()
	case STRING_CONST:
		c.processStringConst()
	case IDENTIFIER:
		c.processIdentifier()
		switch c.tokenizer.Symbol() {
		case '[':
			c.processSymbol('[')
			c.CompileExpression()
			c.processSymbol(']')
		case '(':
			c.processSymbol('(')
			c.CompileExpressionList()
			c.processSymbol(')')
		case '.':
			c.processSymbol('.')
			c.processIdentifier()
			c.processSymbol('(')
			c.CompileExpressionList()
			c.processSymbol(')')
		}
	case KEYWORD:
		switch c.tokenizer.KeyWord() {
		case TRUE:
			c.processKeyword("true")
		case FALSE:
			c.processKeyword("false")
		case NULL:
			c.processKeyword("null")
		case THIS:
			c.processKeyword("this")
		default:
			log.Fatalf("Syntax error: CompileTerm: %s", c.tokenizer.current)
		}
	case SYMBOL:
		switch c.tokenizer.Symbol() {
		case '(':
			c.processSymbol('(')
			c.CompileExpression()
			c.processSymbol(')')
		case '-':
			c.processSymbol('-')
			c.CompileTerm()
		case '~':
			c.processSymbol('~')
			c.CompileTerm()
		default:
			log.Fatalf("Syntax error: CompileTerm: %s", c.tokenizer.current)
		}
	default:
		log.Fatalf("Syntax error: CompileTerm: %s", c.tokenizer.current)
	}

	c.indent = c.indent[2:]
	c.write("</term>\n")
}

func (c *CompilationEngine) CompileExpressionList() int {
	sum := 0

	c.write("<expressionList>\n")
	c.indent = c.indent + "  "

	for c.tokenizer.TokenType() != SYMBOL || c.tokenizer.Symbol() != ')' {
		sum += 1
		c.CompileExpression()

		if c.tokenizer.Symbol() == ',' {
			c.processSymbol(',')
		}
	}

	c.indent = c.indent[2:]
	c.write("</expressionList>\n")

	return sum
}

func (c *CompilationEngine) compileType() {
	switch {
	case c.tokenizer.TokenType() == KEYWORD && c.tokenizer.KeyWord() == INT:
		c.processKeyword("int")
	case c.tokenizer.TokenType() == KEYWORD && c.tokenizer.KeyWord() == CHAR:
		c.processKeyword("char")
	case c.tokenizer.TokenType() == KEYWORD && c.tokenizer.KeyWord() == BOOLEAN:
		c.processKeyword("boolean")
	case c.tokenizer.TokenType() == IDENTIFIER:
		c.processIdentifier()
	default:
		log.Fatalf("Syntax error: compileType: %s", c.tokenizer.current)
	}
}

func (c *CompilationEngine) processKeyword(keyword string) {
	if _, ok := keywordMap[keyword]; c.tokenizer.TokenType() != KEYWORD || !ok || keywordMap[keyword] != c.tokenizer.KeyWord() {
		log.Fatalf("Syntax error: processKeyword: %s", c.tokenizer.current)
	}

	c.write("<keyword> " + keyword + " </keyword>\n")
	c.tokenizer.Advance()
}

func (c *CompilationEngine) processSymbol(symbol rune) {
	if c.tokenizer.TokenType() != SYMBOL {
		log.Fatalf("Syntax error: processSymbol: TokenType is not SYMBOL: %s %d", c.tokenizer.current, c.tokenizer.TokenType())
	}

	if symbol != c.tokenizer.Symbol() {
		log.Fatalf("Syntax error: processSymbol: Symbol received is not the same with current symbol: %s %c", c.tokenizer.current, symbol)
	}

	var output string

	switch symbol {
	case '<':
		output = "&lt;"
	case '>':
		output = "&gt;"
	case '"':
		output = "&quot;"
	case '&':
		output = "&amp;"
	default:
		output = string(symbol)
	}

	c.write("<symbol> " + output + " </symbol>\n")
	c.tokenizer.Advance()
}

func (c *CompilationEngine) processIdentifier() {
	if c.tokenizer.TokenType() != IDENTIFIER {
		log.Fatalf("Syntax error: processIdentifier: %s", c.tokenizer.current)
	}

	c.write("<identifier> " + c.tokenizer.Identifier() + " </identifier>\n")
	c.tokenizer.Advance()
}

func (c *CompilationEngine) processIntConst() {
	if c.tokenizer.TokenType() != INT_CONST {
		log.Fatalf("Syntax error: processIntConst: %s", c.tokenizer.current)
	}

	c.write("<integerConstant> " + strconv.FormatInt(int64(c.tokenizer.IntVal()), 10) + " </integerConstant>\n")
	c.tokenizer.Advance()
}

func (c *CompilationEngine) processStringConst() {
	if c.tokenizer.TokenType() != STRING_CONST {
		log.Fatalf("Syntax error: processIntConst: %s", c.tokenizer.current)
	}

	c.write("<stringConstant> " + c.tokenizer.StringVal() + " </stringConstant>\n")
	c.tokenizer.Advance()
}

func (c *CompilationEngine) write(output string) {
	if _, err := c.writer.WriteString(c.indent + output); err != nil {
		log.Fatalf("Error writing value: %s: %w", output, err)
	}

	if err := c.writer.Flush(); err != nil {
		log.Fatalf("Error writing to file: %w", err)
	}
}
