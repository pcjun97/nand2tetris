package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("usage: hack-assembler file")
	}

	inputFile := os.Args[1]

	if !strings.HasSuffix(inputFile, ".asm") {
		log.Fatalln("invalid file type")
	}

	outputFile := strings.TrimSuffix(inputFile, ".asm") + ".hack"

	p1 := NewParser(inputFile)
	s := NewSymbol()

	i := 0
	for p1.HasMoreLines() {
		p1.Advance()

		instructionType, err := p1.InstructionType()
		if err != nil {
			log.Fatal(err)
		}

		if instructionType == LInstruction {
			symbol, err := p1.Symbol()
			if err != nil {
				log.Fatal(err)
			}

			if s.Contains(symbol) {
				log.Fatal(errors.New("Same label defined at multiple locations"))
			}

			s.AddEntry(symbol, i)
		} else {
			i++
		}
	}

	p2 := NewParser(inputFile)
	c := NewCode()

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	writer := bufio.NewWriter(f)

	i = 16
	for p2.HasMoreLines() {
		p2.Advance()

		instructionType, err := p2.InstructionType()
		if err != nil {
			log.Fatal(err)
		}

		if instructionType == AInstruction {
			symbol, _ := p2.Symbol()

			value, err := strconv.ParseInt(symbol, 10, 0)
			if err != nil {
				if !s.Contains(symbol) {
					s.AddEntry(symbol, i)
					i++
				}
				_value, _ := s.GetAddress(symbol)
				value = int64(_value)
			}

			if value > 32767 {
				log.Fatal(errors.New("Value larger than 15 bits"))
			}

			fmt.Fprintf(writer, "%016b\n", value)
		} else if instructionType == CInstruction {
			comp, _ := p2.Comp()
			dest, _ := p2.Dest()
			jump, _ := p2.Jump()

			compCode, _ := c.Comp(comp)
			destCode, _ := c.Dest(dest)
			jumpCode, _ := c.Jump(jump)

			fmt.Fprintf(writer, "111%s%s%s\n", compCode, destCode, jumpCode)
		}
	}

	if err = writer.Flush(); err != nil {
		log.Fatal(err)
	}
}
