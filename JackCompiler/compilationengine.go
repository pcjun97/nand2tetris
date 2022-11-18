package main

import (
	"os"
	"strconv"
	"strings"
)

type CompilationEngine struct {
	tokenizer      *Tokenizer
	writer         *VMWriter
	cst            *SymbolTable
	sst            *SymbolTable
	className      string
	subroutineName string
	subroutineType KeyWord
	ifCount        int
	whileCount     int
}

func NewCompilationEngine(input, output *os.File) *CompilationEngine {
	tokenizer := NewTokenizer(input)
	tokenizer.Advance()

	writer := NewVMWriter(output)

	c := CompilationEngine{
		tokenizer:  tokenizer,
		writer:     writer,
		cst:        NewSymbolTable(),
		sst:        NewSymbolTable(),
		ifCount:    0,
		whileCount: 0,
	}

	return &c
}

func (c *CompilationEngine) CompileClass() {
	c.cst.Reset()
	c.processKeyword("class")
	c.className = c.processIdentifier()
	c.processSymbol('{')

	for c.tokenizer.TokenType() == TOKEN_TYPE_KEYWORD {
		switch c.tokenizer.KeyWord() {
		case KEYWORD_STATIC, KEYWORD_FIELD:
			c.CompileClassVarDec()
		case KEYWORD_CONSTRUCTOR, KEYWORD_FUNCTION, KEYWORD_METHOD:
			c.CompileSubroutine()
		}
	}

	c.processSymbol('}')
}

func (c *CompilationEngine) CompileClassVarDec() {
	var kind SymbolKind

	switch c.tokenizer.KeyWord() {
	case KEYWORD_STATIC:
		kind = SYMBOL_STATIC
		c.processKeyword("static")
	case KEYWORD_FIELD:
		kind = SYMBOL_FIELD
		c.processKeyword("field")
	default:
		panic("invalid keyword: " + c.tokenizer.current)
	}

	varType := c.processType()

	for {
		varName := c.processIdentifier()
		c.cst.Define(varName, varType, kind)

		if c.tokenizer.Symbol() == ';' {
			break
		}

		c.processSymbol(',')
	}

	c.processSymbol(';')
}

func (c *CompilationEngine) CompileSubroutine() {
	c.sst.Reset()

	keyword := c.tokenizer.KeyWord()
	c.processKeyword("")

	switch keyword {
	case KEYWORD_CONSTRUCTOR, KEYWORD_FUNCTION, KEYWORD_METHOD:
		c.subroutineType = keyword
		if keyword == KEYWORD_METHOD {
			c.sst.Define(KEYWORD_THIS.String(), c.className, SYMBOL_ARG)
		}

	default:
		panic("invalid keyword: " + c.tokenizer.current)
	}

	c.processType()
	c.subroutineName = c.processIdentifier()

	c.processSymbol('(')
	c.CompileParameterList()
	c.processSymbol(')')

	c.CompileSubroutineBody()
}

func (c *CompilationEngine) CompileParameterList() {
	for c.tokenizer.TokenType() != TOKEN_TYPE_SYMBOL {
		argType := c.processType()
		argName := c.processIdentifier()
		c.sst.Define(argName, argType, SYMBOL_ARG)

		if c.tokenizer.Symbol() != ',' {
			break
		}

		c.processSymbol(',')
	}
}

func (c *CompilationEngine) CompileSubroutineBody() {
	c.processSymbol('{')

	for c.tokenizer.TokenType() == TOKEN_TYPE_KEYWORD && c.tokenizer.KeyWord() == KEYWORD_VAR {
		c.CompileVarDec()
	}

	c.writer.WriteFunction(c.className+"."+c.subroutineName, c.sst.VarCount(SYMBOL_VAR))

	if c.subroutineType == KEYWORD_CONSTRUCTOR {
		c.writer.WritePush(STACK_SEGMENT_CONSTANT, c.cst.VarCount(SYMBOL_FIELD))
		c.writer.WriteCall("Memory.alloc", 1)
		c.writer.WritePop(STACK_SEGMENT_POINTER, 0)
	}

	if c.subroutineType == KEYWORD_METHOD {
		c.writer.WritePush(STACK_SEGMENT_ARGUMENT, 0)
		c.writer.WritePop(STACK_SEGMENT_POINTER, 0)
	}

	c.CompileStatements()
	c.processSymbol('}')
}

func (c *CompilationEngine) CompileVarDec() {
	c.processKeyword("var")
	varType := c.processType()

	for {
		identifier := c.processIdentifier()
		c.sst.Define(identifier, varType, SYMBOL_VAR)

		if c.tokenizer.Symbol() == ';' {
			break
		}

		c.processSymbol(',')
	}

	c.processSymbol(';')
}

func (c *CompilationEngine) CompileStatements() {
	for c.tokenizer.TokenType() == TOKEN_TYPE_KEYWORD {
		switch c.tokenizer.KeyWord() {
		case KEYWORD_LET:
			c.CompileLet()
		case KEYWORD_IF:
			c.CompileIf()
		case KEYWORD_WHILE:
			c.CompileWhile()
		case KEYWORD_DO:
			c.CompileDo()
		case KEYWORD_RETURN:
			c.CompileReturn()
		default:
			panic("invalid keyword: " + c.tokenizer.current)
		}
	}
}

func (c *CompilationEngine) CompileLet() {
	c.processKeyword("let")
	kind, index, _ := c.lookupVar(c.processIdentifier())
	if kind == SYMBOL_NONE {
		panic("missing symbol")
	}

	if c.tokenizer.Symbol() == '[' {
		c.writer.WritePush(symbolKindToStackSegment[kind], index)
		c.processSymbol('[')
		c.CompileExpression()
		c.processSymbol(']')
		c.writer.WriteArithmetic(ARITHMETIC_COMMAND_ADD)

		c.processSymbol('=')
		c.CompileExpression()

		c.writer.WritePop(STACK_SEGMENT_TEMP, 0)
		c.writer.WritePop(STACK_SEGMENT_POINTER, 1)
		c.writer.WritePush(STACK_SEGMENT_TEMP, 0)
		c.writer.WritePop(STACK_SEGMENT_THAT, 0)
	} else {
		c.processSymbol('=')
		c.CompileExpression()
		c.writer.WritePop(symbolKindToStackSegment[kind], index)
	}

	c.processSymbol(';')
}

func (c *CompilationEngine) CompileIf() {
	labelElse := "ELSE_" + strconv.Itoa(c.ifCount)
	labelEnd := "IF_END_" + strconv.Itoa(c.ifCount)
	c.ifCount++

	c.processKeyword("if")
	c.processSymbol('(')
	c.CompileExpression()
	c.processSymbol(')')

	c.writer.WriteArithmetic(ARITHMETIC_COMMAND_NOT)
	c.writer.WriteIf(labelElse)

	c.processSymbol('{')
	c.CompileStatements()
	c.processSymbol('}')

	c.writer.WriteGoto(labelEnd)
	c.writer.WriteLabel(labelElse)

	if c.tokenizer.TokenType() == TOKEN_TYPE_KEYWORD && c.tokenizer.KeyWord() == KEYWORD_ELSE {
		c.processKeyword("else")
		c.processSymbol('{')
		c.CompileStatements()
		c.processSymbol('}')
	}

	c.writer.WriteLabel(labelEnd)
}

func (c *CompilationEngine) CompileWhile() {
	labelStart := "WHILE_START_" + strconv.Itoa(c.whileCount)
	labelEnd := "WHILE_END_" + strconv.Itoa(c.whileCount)
	c.whileCount++

	c.processKeyword("while")
	c.writer.WriteLabel(labelStart)

	c.processSymbol('(')
	c.CompileExpression()
	c.processSymbol(')')

	c.writer.WriteArithmetic(ARITHMETIC_COMMAND_NOT)
	c.writer.WriteIf(labelEnd)

	c.processSymbol('{')
	c.CompileStatements()
	c.processSymbol('}')

	c.writer.WriteGoto(labelStart)
	c.writer.WriteLabel(labelEnd)
}

func (c *CompilationEngine) CompileDo() {
	c.processKeyword("do")
	c.CompileExpression()
	c.writer.WritePop(STACK_SEGMENT_TEMP, 0)
	c.processSymbol(';')
}

func (c *CompilationEngine) CompileReturn() {
	c.processKeyword("return")
	if c.tokenizer.TokenType() != TOKEN_TYPE_SYMBOL || c.tokenizer.Symbol() != ';' {
		c.CompileExpression()
	}
	c.writer.WriteReturn()
	c.processSymbol(';')
}

func (c *CompilationEngine) CompileExpression() {
	c.CompileTerm()

	for strings.ContainsRune("+-*/&|<>=", c.tokenizer.Symbol()) {
		symbol := c.processSymbol(-1)
		c.CompileTerm()

		switch symbol {
		case '+':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_ADD)
		case '-':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_SUB)
		case '*':
			c.writer.WriteCall("Math.multiply", 2)
		case '/':
			c.writer.WriteCall("Math.divide", 2)
		case '&':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_AND)
		case '|':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_OR)
		case '<':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_LT)
		case '>':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_GT)
		case '=':
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_EQ)
		}
	}
}

func (c *CompilationEngine) CompileTerm() {
	switch c.tokenizer.TokenType() {
	case TOKEN_TYPE_INT_CONST:
		val := c.processIntConst()
		c.writer.WritePush(STACK_SEGMENT_CONSTANT, val)

	case TOKEN_TYPE_STRING_CONST:
		val := c.processStringConst()
		c.writer.WritePush(STACK_SEGMENT_CONSTANT, len(val))
		c.writer.WriteCall("String.new", 1)
		for _, char := range val {
			c.writer.WritePush(STACK_SEGMENT_CONSTANT, int(char))
			c.writer.WriteCall("String.appendChar", 2)
		}

	case TOKEN_TYPE_IDENTIFIER:
		identifier := c.processIdentifier()

		switch c.tokenizer.Symbol() {
		case '[':
			kind, index, _ := c.lookupVar(identifier)
			if kind == SYMBOL_NONE {
				panic("missing symbol")
			}
			c.writer.WritePush(symbolKindToStackSegment[kind], index)
			c.processSymbol('[')
			c.CompileExpression()
			c.processSymbol(']')
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_ADD)
			c.writer.WritePop(STACK_SEGMENT_POINTER, 1)
			c.writer.WritePush(STACK_SEGMENT_THAT, 0)

		case '(':
			c.compileMethodCall("this", identifier)

		case '.':
			c.processSymbol('.')
			subroutineName := c.processIdentifier()
			kind, _, _ := c.lookupVar(identifier)
			if kind == SYMBOL_NONE {
				c.compileFunctionCall(identifier, subroutineName)
			} else {
				c.compileMethodCall(identifier, subroutineName)
			}

		default:
			kind, index, _ := c.lookupVar(identifier)
			if kind == SYMBOL_NONE {
				panic("missing symbol")
			}
			c.writer.WritePush(symbolKindToStackSegment[kind], index)
		}

	case TOKEN_TYPE_KEYWORD:
		switch c.tokenizer.KeyWord() {
		case KEYWORD_TRUE:
			c.processKeyword("true")
			c.writer.WritePush(STACK_SEGMENT_CONSTANT, 1)
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_NEG)

		case KEYWORD_FALSE, KEYWORD_NULL:
			c.processKeyword("")
			c.writer.WritePush(STACK_SEGMENT_CONSTANT, 0)

		case KEYWORD_THIS:
			c.processKeyword("this")
			kind, index, _ := c.lookupVar(KEYWORD_THIS.String())
			if kind != SYMBOL_NONE {
				c.writer.WritePush(symbolKindToStackSegment[kind], index)
			} else {
				c.writer.WritePush(STACK_SEGMENT_POINTER, 0)
			}

		default:
			panic("invalid keyword: " + c.tokenizer.current)
		}

	case TOKEN_TYPE_SYMBOL:
		switch c.tokenizer.Symbol() {
		case '(':
			c.processSymbol('(')
			c.CompileExpression()
			c.processSymbol(')')

		case '-':
			c.processSymbol('-')
			c.CompileTerm()
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_NEG)

		case '~':
			c.processSymbol('~')
			c.CompileTerm()
			c.writer.WriteArithmetic(ARITHMETIC_COMMAND_NOT)

		default:
			panic("invalid symbol: " + c.tokenizer.current)
		}

	default:
		panic("invalid token: " + c.tokenizer.current)
	}
}

func (c *CompilationEngine) CompileExpressionList() int {
	sum := 0

	for c.tokenizer.TokenType() != TOKEN_TYPE_SYMBOL || c.tokenizer.Symbol() != ')' {
		c.CompileExpression()
		sum++

		if c.tokenizer.Symbol() == ',' {
			c.processSymbol(',')
		}
	}

	return sum
}

func (c *CompilationEngine) compileFunctionCall(className, functionName string) {
	c.processSymbol('(')
	nArgs := c.CompileExpressionList()
	c.processSymbol(')')
	c.writer.WriteCall(className+"."+functionName, nArgs)
}

func (c *CompilationEngine) compileMethodCall(object, methodName string) {
	kind, index, className := c.lookupVar(object)
	switch {
	case kind != SYMBOL_NONE:
		c.writer.WritePush(symbolKindToStackSegment[kind], index)
	case object == "this":
		c.writer.WritePush(STACK_SEGMENT_POINTER, 0)
		className = c.className
	default:
		panic("missing symbol: " + object)
	}

	c.processSymbol('(')
	nArgs := c.CompileExpressionList() + 1
	c.processSymbol(')')
	c.writer.WriteCall(className+"."+methodName, nArgs)
}

func (c *CompilationEngine) processType() string {
	var result string

	switch c.tokenizer.TokenType() {
	case TOKEN_TYPE_KEYWORD:
		switch c.processKeyword("") {
		case KEYWORD_INT:
			result = KEYWORD_INT.String()
		case KEYWORD_CHAR:
			result = KEYWORD_CHAR.String()
		case KEYWORD_BOOLEAN:
			result = KEYWORD_INT.String()
		case KEYWORD_VOID:
			result = KEYWORD_VOID.String()
		default:
			panic("invalid keyword")
		}

	case TOKEN_TYPE_IDENTIFIER:
		result = c.processIdentifier()

	default:
		panic("invalid token type")
	}

	return result
}

func (c *CompilationEngine) processKeyword(expected string) KeyWord {
	keyword := c.tokenizer.KeyWord()

	if len(expected) > 0 {
		value, ok := stringToKeyword[expected]
		if !ok {
			panic("invalid keyword")
		}

		if value != keyword {
			panic("mismatch keyword")
		}
	}

	c.tokenizer.Advance()

	return keyword
}

func (c *CompilationEngine) processSymbol(expected rune) rune {
	symbol := c.tokenizer.Symbol()

	if expected >= 0 && expected != symbol {
		panic("symbol mismatch")
	}

	c.tokenizer.Advance()

	return symbol
}

func (c *CompilationEngine) processIdentifier() string {
	identifier := c.tokenizer.Identifier()
	c.tokenizer.Advance()
	return identifier
}

func (c *CompilationEngine) processIntConst() int {
	intVal := c.tokenizer.IntVal()
	c.tokenizer.Advance()
	return intVal
}

func (c *CompilationEngine) processStringConst() string {
	stringVal := c.tokenizer.StringVal()
	c.tokenizer.Advance()
	return stringVal
}

func (c *CompilationEngine) lookupVar(v string) (SymbolKind, int, string) {
	var (
		kind       SymbolKind
		index      int
		symbolType string
	)

	kind = c.sst.KindOf(v)
	if kind == SYMBOL_NONE {
		kind = c.cst.KindOf(v)
		index = c.cst.IndexOf(v)
		symbolType = c.cst.TypeOf(v)
	} else {
		index = c.sst.IndexOf(v)
		symbolType = c.sst.TypeOf(v)
	}

	return kind, index, symbolType
}
