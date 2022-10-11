package main

import (
	"errors"
	"fmt"
)

var compMap map[string]int = map[string]int{
	"0":   0b0101010,
	"1":   0b0111111,
	"-1":  0b0111010,
	"D":   0b0001100,
	"A":   0b0110000,
	"M":   0b1110000,
	"!D":  0b0001101,
	"!A":  0b0110001,
	"!M":  0b1110001,
	"-D":  0b0001111,
	"-A":  0b0110011,
	"-M":  0b1110011,
	"D+1": 0b0011111,
	"A+1": 0b0110111,
	"M+1": 0b1110111,
	"D-1": 0b0001110,
	"A-1": 0b0110010,
	"M-1": 0b1110010,
	"D+A": 0b0000010,
	"D+M": 0b1000010,
	"D-A": 0b0010011,
	"D-M": 0b1010011,
	"A-D": 0b0000111,
	"M-D": 0b1000111,
	"D&A": 0b0000000,
	"D&M": 0b1000000,
	"D|A": 0b0010101,
	"D|M": 0b1010101,
}

var destMap map[string]int = map[string]int{
	"null": 0b000,
	"M":    0b001,
	"D":    0b010,
	"DM":   0b011,
	"MD":   0b011,
	"A":    0b100,
	"AM":   0b101,
	"MA":   0b101,
	"AD":   0b110,
	"DA":   0b110,
	"ADM":  0b111,
	"AMD":  0b111,
	"DAM":  0b111,
	"DMA":  0b111,
	"MAD":  0b111,
	"MDA":  0b111,
}

var jumpMap map[string]int = map[string]int{
	"null": 0b000,
	"JGT":  0b001,
	"JEQ":  0b010,
	"JGE":  0b011,
	"JLT":  0b100,
	"JNE":  0b101,
	"JLE":  0b110,
	"JMP":  0b111,
}

type Code struct{}

func NewCode() *Code {
	c := Code{}
	return &c
}

func (c *Code) Dest(input string) (string, error) {
	value, ok := destMap[input]
	if ok {
		return fmt.Sprintf("%03b", value), nil
	}
	return "", errors.New("Invalid dest")
}

func (c *Code) Comp(input string) (string, error) {
	value, ok := compMap[input]
	if ok {
		return fmt.Sprintf("%07b", value), nil
	}
	return "", errors.New("Invalid comp")
}

func (c *Code) Jump(input string) (string, error) {
	value, ok := jumpMap[input]
	if ok {
		return fmt.Sprintf("%03b", value), nil
	}
	return "", errors.New("Invalid jump")
}
