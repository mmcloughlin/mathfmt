package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var write = flag.Bool("w", false, "write result to (source) file instead of stdout")

func main() {
	log.SetPrefix("mathfmt: ")
	log.SetFlags(0)

	flag.Parse()

	for _, path := range flag.Args() {
		processPath(path)
	}
}

func processPath(path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.IsDir() {
		entries, err := os.ReadDir(path)

		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range entries {
			processPath(entry.Name())
		}

	} else {
		process(path)
	}
}

func process(filename string) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	b, err := Format(src)
	if err != nil {
		log.Fatal(err)
	}

	if *write {
		err = ioutil.WriteFile(filename, b, 0o644)
	} else {
		_, err = os.Stdout.Write(b)
	}

	if err != nil {
		log.Fatal(err)
	}
}
