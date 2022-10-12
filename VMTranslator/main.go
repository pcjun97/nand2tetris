package main

import (
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("usage: VMTranslator source")
	}

	inputPath := os.Args[1]

	pathInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var outputFile string
	var inputFiles []string

	if pathInfo.IsDir() {
		files, err := os.ReadDir(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".vm") {
				inputFiles = append(inputFiles, path.Join(inputPath, file.Name()))
			}
		}

		outputFile = path.Join(inputPath, path.Base(inputPath)+".asm")
	} else {
		if !strings.HasSuffix(inputPath, ".vm") {
			log.Fatalln("invalid file type")
		}

		outputFile = strings.TrimSuffix(inputPath, ".vm") + ".asm"
		inputFiles = append(inputFiles, inputPath)
	}

	c := NewCodeWriter(outputFile)

	for _, file := range inputFiles {
		p := NewParser(file)
		c.SetFileName(file)

		for p.HasMoreLines() {
			p.Advance()
			switch p.CommandType() {
			case C_ARITHMETIC:
				c.WriteArithmetic(p.Arg1())
			case C_PUSH:
				c.WritePushPop("push", p.Arg1(), p.Arg2())
			case C_POP:
				c.WritePushPop("pop", p.Arg1(), p.Arg2())
			case C_LABEL:
				c.WriteLabel(p.Arg1())
			case C_GOTO:
				c.WriteGoto(p.Arg1())
			case C_IF:
				c.WriteIf(p.Arg1())
			case C_FUNCTION:
				c.WriteFunction(p.Arg1(), p.Arg2())
			case C_CALL:
				c.WriteCall(p.Arg1(), p.Arg2())
			case C_RETURN:
				c.WriteReturn()
			}
		}
	}

	c.Close()
}
