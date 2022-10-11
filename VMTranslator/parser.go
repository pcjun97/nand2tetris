package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type CommandType int

const (
	C_ERROR CommandType = iota
	C_ARITHMETIC
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

var arithmeticCommands map[string]bool = map[string]bool{
	"add": true,
	"sub": true,
	"neg": true,
	"eq":  true,
	"gt":  true,
	"lt":  true,
	"and": true,
	"or":  true,
	"not": true,
}

type Parser struct {
	file        *os.File
	scanner     *bufio.Scanner
	next        string
	current     string
	commandType CommandType
	fields      []string
}

func NewParser(file string) *Parser {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	p := Parser{
		file:        f,
		scanner:     scanner,
		next:        "init",
		current:     "",
		commandType: C_ERROR,
	}

	p.Advance()
	p.current = ""

	return &p
}

func (p *Parser) HasMoreLines() bool {
	return len(p.next) > 0
}

func (p *Parser) Advance() {
	if p.HasMoreLines() {
		p.current = p.next
		p.fields = strings.Fields(p.current)
		p.setCommandType()

		for {
			if ok := p.scanner.Scan(); !ok {
				err := p.scanner.Err()
				if err != nil {
					log.Fatal(err)
				}

				err = p.file.Close()
				if err != nil {
					log.Fatal(err)
				}

				p.next = ""
				break
			}

			line := p.scanner.Text()
			if comment := strings.Index(line, "//"); comment >= 0 {
				line = line[:comment]
			}
			line = strings.TrimSpace(line)

			if len(line) > 0 {
				p.next = line
				break
			}
		}
	}
}

func (p *Parser) setCommandType() {
	switch {
	case p.fields[0] == "push" && len(p.fields) == 3:
		p.commandType = C_PUSH
	case p.fields[0] == "pop" && len(p.fields) == 3:
		p.commandType = C_POP
	default:
		_, ok := arithmeticCommands[p.fields[0]]
		if ok && len(p.fields) == 1 {
			p.commandType = C_ARITHMETIC
		} else {
			p.commandType = C_ERROR
		}
	}
}

func (p *Parser) CommandType() CommandType {
	if p.commandType == C_ERROR {
		log.Fatalf("Unknown command: %s\n", p.fields[0])
	}
	return p.commandType
}

func (p *Parser) Arg1() string {
	if p.commandType == C_ERROR || p.commandType == C_RETURN {
		return ""
	}

	if p.commandType == C_ARITHMETIC {
		return p.fields[0]
	}

	return p.fields[1]
}

func (p *Parser) Arg2() int {
	if p.commandType == C_PUSH || p.commandType == C_POP || p.commandType == C_FUNCTION || p.commandType == C_CALL {
		value, err := strconv.ParseInt(p.fields[2], 10, 0)
		if err != nil {
			return -1
		}

		return int(value)
	}

	return -1
}
