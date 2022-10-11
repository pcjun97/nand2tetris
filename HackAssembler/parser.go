package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type InstructionType int

const (
	ErrorInstruction InstructionType = iota
	AInstruction
	CInstruction
	LInstruction
)

type Parser struct {
	file    *os.File
	scanner *bufio.Scanner
	next    string
	current string
}

func NewParser(file string) *Parser {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	p := Parser{
		file:    f,
		scanner: scanner,
		next:    "init",
		current: "",
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

func (p *Parser) InstructionType() (InstructionType, error) {
	symbol := regexp.MustCompile(`^[a-zA-Z0-9_.$:][a-zA-Z0-9_.$:]*$`)

	if p.current[0] == '@' {
		if symbol.MatchString(p.current[1:]) {
			return AInstruction, nil
		}
		_, err := strconv.ParseInt(p.current[1:], 10, 15)
		if err != nil {
			log.Fatal(err)
		}
		return AInstruction, nil
	}

	if p.current[0] == '(' && p.current[len(p.current)-1] == ')' && symbol.MatchString(p.current[1:len(p.current)-1]) {
		return LInstruction, nil
	}

	i := strings.Index(p.current, "=")
	j := strings.Index(p.current, ";")
	if i > 0 || j > 0 {
		comp := ""
		dest := "null"
		jump := "null"

		if i > 0 && j > 0 {
			dest = p.current[:i]
			comp = p.current[i+1 : j]
			jump = p.current[j+1:]
		} else if i > 0 {
			dest = p.current[:i]
			comp = p.current[i+1:]
		} else {
			comp = p.current[:j]
			jump = p.current[j+1:]
		}

		_, compValid := compMap[comp]
		_, destValid := destMap[dest]
		_, jumpValid := jumpMap[jump]

		if compValid && destValid && jumpValid {
			return CInstruction, nil
		}
	}

	return ErrorInstruction, fmt.Errorf("Invalid instruction format: %s\n", p.current)
}

func (p *Parser) Symbol() (string, error) {
	instructionType, err := p.InstructionType()
	if err != nil {
		log.Fatal(err)
	}

	if instructionType == AInstruction {
		return p.current[1:], nil
	} else if instructionType == LInstruction {
		return p.current[1 : len(p.current)-1], nil
	}

	return "", errors.New("No valid symbol found")
}

func (p *Parser) Dest() (string, error) {
	instructionType, err := p.InstructionType()
	if err != nil {
		log.Fatal(err)
	}

	if instructionType == CInstruction {
		i := strings.Index(p.current, "=")
		if i > 0 {
			return p.current[:i], nil
		}
		return "null", nil
	}

	return "", errors.New("No valid dest found")
}

func (p *Parser) Comp() (string, error) {
	instructionType, err := p.InstructionType()
	if err != nil {
		log.Fatal(err)
	}

	if instructionType == CInstruction {
		i := strings.Index(p.current, "=")
		j := strings.Index(p.current, ";")

		if i > 0 && j > 0 {
			return p.current[i+1 : j], nil
		} else if i > 0 {
			return p.current[i+1:], nil
		} else if j > 0 {
			return p.current[:j], nil
		}
	}

	return "", errors.New("No valid comp found")
}

func (p *Parser) Jump() (string, error) {
	instructionType, err := p.InstructionType()
	if err != nil {
		log.Fatal(err)
	}

	if instructionType == CInstruction {
		if j := strings.Index(p.current, ";"); j > 0 {
			return p.current[j+1:], nil
		}
		return "null", nil
	}

	return "", errors.New("No valid jump found")
}
