package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed embeds/init.asm
var initAsm string

//go:embed embeds/popD.asm
var popDAsm string

//go:embed embeds/fetchM.asm
var fetchMAsm string

//go:embed embeds/conditional.asm
var conditionalAsm string

//go:embed embeds/push.asm
var pushAsm string

//go:embed embeds/pop.asm
var popAsm string

//go:embed embeds/pushAddress.asm
var pushAddressAsm string

//go:embed embeds/pushSymbol.asm
var pushSymbolAsm string

//go:embed embeds/popSymbol.asm
var popSymbolAsm string

//go:embed embeds/pushConstant.asm
var pushConstantAsm string

//go:embed embeds/endLoop.asm
var endLoopAsm string

//go:embed embeds/goto.asm
var gotoAsm string

//go:embed embeds/ifgoto.asm
var ifgotoAsm string

//go:embed embeds/function.asm
var functionAsm string

//go:embed embeds/call.asm
var callAsm string

//go:embed embeds/return.asm
var returnAsm string

var segmentMapping map[string]string = map[string]string{
	"argument": "ARG",
	"local":    "LCL",
	"this":     "THIS",
	"that":     "THAT",
	"temp":     "TEMP",
	"static":   "LCL",
	"constant": "LCL",
	"pointer":  "LCL",
}

type CodeWriter struct {
	file     *os.File
	writer   *bufio.Writer
	id       map[string]int
	fileName string
	function string
}

func NewCodeWriter(file string) *CodeWriter {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}

	writer := bufio.NewWriter(f)
	id := make(map[string]int)

	c := CodeWriter{
		file:   f,
		writer: writer,
		id:     id,
	}

	c.write(initAsm)
	c.WriteCall("Sys.init", 0)

	return &c
}

func (c *CodeWriter) SetFileName(fileName string) {
	c.fileName = strings.TrimSuffix(path.Base(fileName), ".vm")
}

func (c *CodeWriter) WriteArithmetic(command string) {
	output := ""
	comment := fmt.Sprintf("// %s\n", command)

	switch command {
	case "add":
		output = comment + popDAsm + fetchMAsm + "M=M+D\n"

	case "sub":
		output = comment + popDAsm + fetchMAsm + "M=M-D\n"

	case "and":
		output = comment + popDAsm + fetchMAsm + "M=M&D\n"

	case "or":
		output = comment + popDAsm + fetchMAsm + "M=M|D\n"

	case "neg":
		output = comment + fetchMAsm + "M=-M\n"

	case "not":
		output = comment + fetchMAsm + "M=!M\n"

	case "eq", "gt", "lt":
		label := fmt.Sprintf("%s_%d", strings.ToUpper(command), c.getId(c.function+"$"+command))
		if c.function != "" {
			label = c.function + "$" + label
		}

		output = comment + popDAsm + fetchMAsm + fmt.Sprintf(conditionalAsm, strings.ToUpper(command), label)
	}

	c.write(output)
}

func (c *CodeWriter) WritePushPop(command, segment string, index int) {
	output := ""
	comment := fmt.Sprintf("// %s %s %d\n", command, segment, index)

	switch segment {
	case "pointer":
		output = comment + c.getPointer(command, index)
	case "temp":
		output = comment + c.getTemp(command, index)
	case "constant":
		output = comment + c.getConstant(command, index)
	case "static":
		output = comment + c.getStatic(command, index)
	default:
		output = comment + c.getDefaultPushPop(command, segment, index)
	}

	c.write(output)
}

func (c *CodeWriter) WriteLabel(label string) {
	comment := fmt.Sprintf("// label %s\n", label)

	if c.function != "" {
		label = c.function + "$" + label
	}

	output := fmt.Sprintf("(%s)\n", label)
	c.write(comment + output)
}

func (c *CodeWriter) WriteGoto(label string) {
	comment := fmt.Sprintf("// goto %s\n", label)

	if c.function != "" {
		label = c.function + "$" + label
	}

	output := fmt.Sprintf(gotoAsm, label)
	c.write(comment + output)
}

func (c *CodeWriter) WriteIf(label string) {
	comment := fmt.Sprintf("// if-goto %s\n", label)

	if c.function != "" {
		label = c.function + "$" + label
	}

	output := popDAsm + fmt.Sprintf(ifgotoAsm, label)
	c.write(comment + output)
}

func (c *CodeWriter) WriteFunction(label string, nVars int) {
	comment := fmt.Sprintf("// function %s %d\n", label, nVars)
	output := fmt.Sprintf(functionAsm, label, nVars)
	c.write(comment + output)
	c.function = label
}

func (c *CodeWriter) WriteCall(label string, nArgs int) {
	comment := fmt.Sprintf("// call %s %d\n", label, nArgs)
	returnAddress := fmt.Sprintf("%s$ret%d", c.function, c.getId(c.function+"$ret"))
	output := fmt.Sprintf(pushAddressAsm, returnAddress) +
		fmt.Sprintf(pushSymbolAsm, "LCL") +
		fmt.Sprintf(pushSymbolAsm, "ARG") +
		fmt.Sprintf(pushSymbolAsm, "THIS") +
		fmt.Sprintf(pushSymbolAsm, "THAT") +
		fmt.Sprintf(callAsm, label, nArgs, returnAddress)
	c.write(comment + output)
}

func (c *CodeWriter) WriteReturn() {
	comment := fmt.Sprintf("// return\n")
	c.write(comment + returnAsm)
}

func (c *CodeWriter) Close() {
	c.write(endLoopAsm)

	err := c.writer.Flush()
	if err != nil {
		log.Fatal(err)
	}

	err = c.file.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *CodeWriter) write(output string) {
	c.writer.WriteString(output)
	if err := c.writer.Flush(); err != nil {
		log.Fatal(err)
	}
}

func (c *CodeWriter) getId(key string) int {
	id, ok := c.id[key]
	if !ok {
		c.id[key] = 0
	}

	c.id[key] = id + 1
	return id
}

func (c *CodeWriter) getDefaultPushPop(command, segment string, index int) string {
	var output string
	v, _ := segmentMapping[segment]

	switch command {
	case "push":
		output = fmt.Sprintf(pushAsm, v, index)
	case "pop":
		output = fmt.Sprintf(popAsm, v, index)
	}

	return output
}

func (c *CodeWriter) getPointer(command string, index int) string {
	var output string

	switch {
	case command == "push" && index == 0:
		output = fmt.Sprintf(pushSymbolAsm, "THIS")
	case command == "push" && index == 1:
		output = fmt.Sprintf(pushSymbolAsm, "THAT")
	case command == "pop" && index == 0:
		output = fmt.Sprintf(popSymbolAsm, "THIS")
	case command == "pop" && index == 1:
		output = fmt.Sprintf(popSymbolAsm, "THAT")
	}

	return output
}

func (c *CodeWriter) getTemp(command string, index int) string {
	var output string

	register := fmt.Sprintf("R%d", 5+index)

	switch command {
	case "push":
		output = fmt.Sprintf(pushSymbolAsm, register)
	case "pop":
		output = fmt.Sprintf(popSymbolAsm, register)
	}

	return output
}

func (c *CodeWriter) getConstant(command string, index int) string {
	var output string

	switch command {
	case "push":
		output = fmt.Sprintf(pushConstantAsm, index)
	}

	return output
}

func (c *CodeWriter) getStatic(command string, index int) string {
	var output string
	v := c.fileName + "." + strconv.FormatInt(int64(index), 10)

	switch command {
	case "push":
		output = fmt.Sprintf(pushSymbolAsm, v)
	case "pop":
		output = fmt.Sprintf(popSymbolAsm, v)
	}

	return output
}
