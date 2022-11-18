package main

import (
	"bufio"
	"os"
	"strconv"
)

type StackSegment int

const (
	STACK_SEGMENT_UNKNOWN StackSegment = iota
	STACK_SEGMENT_CONSTANT
	STACK_SEGMENT_ARGUMENT
	STACK_SEGMENT_LOCAL
	STACK_SEGMENT_STATIC
	STACK_SEGMENT_THIS
	STACK_SEGMENT_THAT
	STACK_SEGMENT_POINTER
	STACK_SEGMENT_TEMP
)

var stackSegmentNames = []string{
	"",
	"constant",
	"argument",
	"local",
	"static",
	"this",
	"that",
	"pointer",
	"temp",
}

func (s StackSegment) String() string {
	return stackSegmentNames[s]
}

var symbolKindToStackSegment = map[SymbolKind]StackSegment{
	SYMBOL_NONE:   STACK_SEGMENT_UNKNOWN,
	SYMBOL_ARG:    STACK_SEGMENT_ARGUMENT,
	SYMBOL_VAR:    STACK_SEGMENT_LOCAL,
	SYMBOL_STATIC: STACK_SEGMENT_STATIC,
	SYMBOL_FIELD:  STACK_SEGMENT_THIS,
}

type ArithmeticCommand int

const (
	ARITHMETIC_COMMAND_UNKNOWN ArithmeticCommand = iota
	ARITHMETIC_COMMAND_ADD
	ARITHMETIC_COMMAND_SUB
	ARITHMETIC_COMMAND_NEG
	ARITHMETIC_COMMAND_EQ
	ARITHMETIC_COMMAND_GT
	ARITHMETIC_COMMAND_LT
	ARITHMETIC_COMMAND_AND
	ARITHMETIC_COMMAND_OR
	ARITHMETIC_COMMAND_NOT
)

var arithmeticCommandNames = []string{
	"",
	"add",
	"sub",
	"neg",
	"eq",
	"gt",
	"lt",
	"and",
	"or",
	"not",
}

func (a ArithmeticCommand) String() string {
	return arithmeticCommandNames[a]
}

type VMWriter struct {
	writer *bufio.Writer
}

func NewVMWriter(output *os.File) *VMWriter {
	writer := bufio.NewWriter(output)

	v := VMWriter{
		writer: writer,
	}

	return &v
}

func (v *VMWriter) WritePush(segment StackSegment, index int) {
	v.write("push " + segment.String() + " " + strconv.Itoa(index))
}

func (v *VMWriter) WritePop(segment StackSegment, index int) {
	v.write("pop " + segment.String() + " " + strconv.Itoa(index))
}

func (v *VMWriter) WriteArithmetic(command ArithmeticCommand) {
	v.write(command.String())
}

func (v *VMWriter) WriteLabel(label string) {
	v.write("label " + label)
}

func (v *VMWriter) WriteGoto(label string) {
	v.write("goto " + label)
}

func (v *VMWriter) WriteIf(label string) {
	v.write("if-goto " + label)
}

func (v *VMWriter) WriteCall(name string, nArgs int) {
	v.write("call " + name + " " + strconv.Itoa(nArgs))
}

func (v *VMWriter) WriteFunction(name string, nArgs int) {
	v.write("function " + name + " " + strconv.Itoa(nArgs))
}

func (v *VMWriter) WriteReturn() {
	v.write("return")
}

func (v *VMWriter) write(value string) {
	if _, err := v.writer.WriteString(value + "\n"); err != nil {
		panic(err)
	}

	if err := v.writer.Flush(); err != nil {
		panic(err)
	}
}
