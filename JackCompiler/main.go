package main

import (
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("usage: JackCompiler source")
	}

	inputPath := os.Args[1]

	pathInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var paths []string

	if pathInfo.IsDir() {
		files, err := os.ReadDir(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".jack") {
				paths = append(paths, path.Join(inputPath, file.Name()))
			}
		}

	} else {
		if !strings.HasSuffix(inputPath, ".jack") {
			log.Fatalln("invalid file type")
		}

		paths = append(paths, inputPath)
	}

	for _, inputFilePath := range paths {
		inputFile, err := os.Open(inputFilePath)
		if err != nil {
			panic(err)
		}
		defer inputFile.Close()

		outputFilePath := strings.TrimSuffix(inputFilePath, ".jack") + ".vm"
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()

		c := NewCompilationEngine(inputFile, outputFile)
		c.CompileClass()
	}
}
