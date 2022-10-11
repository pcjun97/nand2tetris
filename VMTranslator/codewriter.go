package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed embeds/popD.asm
var popD string

//go:embed embeds/fetchM.asm
var fetchM string

//go:embed embeds/conditional.asm
var conditional string

//go:embed embeds/push.asm
var push string

//go:embed embeds/pop.asm
var pop string

//go:embed embeds/pushPointer.asm
var pushPointer string

//go:embed embeds/popPointer.asm
var popPointer string

//go:embed embeds/pushConstant.asm
var pushConstant string

//go:embed embeds/pushPointer.asm
var pushStatic string

//go:embed embeds/popPointer.asm
var popStatic string

//go:embed embeds/endLoop.asm
var endLoop string

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
}

func NewCodeWriter(file string) *CodeWriter {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}

	writer := bufio.NewWriter(f)

	c := CodeWriter{
		file:   f,
		writer: writer,
	}

	return &c
}

func (c *CodeWriter) SetFileName(fileName string) {
	c.fileName = fileName
}

func (c *CodeWriter) WriteArithmetic(command string) {
	output := ""
	comment := fmt.Sprintf("// %s\n", command)

	switch command {
	case "add":
		output = comment + popD + fetchM + "M=M+D\n"

	case "sub":
		output = comment + popD + fetchM + "M=M-D\n"

	case "and":
		output = comment + popD + fetchM + "M=M&D\n"

	case "or":
		output = comment + popD + fetchM + "M=M|D\n"

	case "neg":
		output = comment + fetchM + "M=-M\n"

	case "not":
		output = comment + fetchM + "M=!M\n"

	case "eq", "gt", "lt":
		output = comment + popD + fetchM + fmt.Sprintf(conditional, strings.ToUpper(command), c.getId(command))
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
}

func (c *CodeWriter) WriteGoto(label string) {
}

func (c *CodeWriter) WriteIf(label string) {
}

func (c *CodeWriter) WriteFunction(label string, nVars int) {
}

func (c *CodeWriter) WriteCall(label string, nArgs int) {
}

func (c *CodeWriter) WriteReturn() {
}

func (c *CodeWriter) Close() {
	c.write(endLoop)

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
		output = fmt.Sprintf(push, v, index)
	case "pop":
		output = fmt.Sprintf(pop, v, index)
	}

	return output
}

func (c *CodeWriter) getPointer(command string, index int) string {
	var output string

	switch {
	case command == "push" && index == 0:
		output = fmt.Sprintf(pushPointer, "THIS")
	case command == "push" && index == 1:
		output = fmt.Sprintf(pushPointer, "THAT")
	case command == "pop" && index == 0:
		output = fmt.Sprintf(popPointer, "THIS")
	case command == "pop" && index == 1:
		output = fmt.Sprintf(popPointer, "THAT")
	}

	return output
}

func (c *CodeWriter) getTemp(command string, index int) string {
	var output string

	switch command {
	case "push":
		output = fmt.Sprintf(push, "TEMP", index+5)
	case "pop":
		output = fmt.Sprintf(pop, "TEMP", index+5)
	}

	return output
}

func (c *CodeWriter) getConstant(command string, index int) string {
	var output string

	switch command {
	case "push":
		output = fmt.Sprintf(pushConstant, index)
	}

	return output
}

func (c *CodeWriter) getStatic(command string, index int) string {
	var output string
	v := c.fileName + "." + strconv.FormatInt(int64(index), 10)

	switch command {
	case "push":
		output = fmt.Sprintf(pushStatic, v)
	case "pop":
		output = fmt.Sprintf(popStatic, v)
	}

	return output
}
