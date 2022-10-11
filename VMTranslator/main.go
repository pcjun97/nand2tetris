package main

import (
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("usage: vm-translator source")
	}

	inputPath := os.Args[1]

	pathInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var outputFile string
	var inputFiles []string

	if pathInfo.IsDir() {
		outputFile = inputPath + path.Base(inputPath) + ".asm"

		files, err := os.ReadDir(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".vm") {
				inputFiles = append(inputFiles, inputPath+file.Name())
			}
		}
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

		for p.HasMoreLines() {
			p.Advance()
			switch p.CommandType() {
			case C_ARITHMETIC:
				c.WriteArithmetic(p.Arg1())
			case C_PUSH:
				c.WritePushPop("push", p.Arg1(), p.Arg2())
			case C_POP:
				c.WritePushPop("pop", p.Arg1(), p.Arg2())
			}
		}
	}

	c.Close()
}
