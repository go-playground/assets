package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-playground/assets"
	"github.com/go-playground/bundler"
)

var (
	flagFileOrDir  = flag.String("i", "", "Directory to bundle files for recursivly.")
	flagLeftDelim  = flag.String("ld", "", "The Left Delimiter for file includes")
	flagRightDelim = flag.String("rd", "", "The Right Delimiter for file includes")
	flagIgnore     = flag.String("ignore", "", "Regexp for files/dirs we should ignore i.e. \\.gitignore; only used when -i option is a DIR")

	input      string
	leftDelim  string
	rightDelim string
	ignore     string

	ignoreRegexp *regexp.Regexp

	processed []*bundler.ProcessedFile
)

func main() {
	parseFlags()

	files := assets.New(leftDelim, rightDelim)
	processed, err := files.Generate(input, ignoreRegexp)
	if err != nil {
		panic(err)
	}

	printResults(processed)
}

func printResults(processed []*bundler.ProcessedFile) {

	fmt.Printf("The following files were processed:\n\n")

	for _, file := range processed {
		fmt.Println("  " + file.NewFilename)
	}

	fmt.Printf("\n\n")
}

func parseFlags() {

	flag.Parse()

	input = strings.TrimSpace(*flagFileOrDir)
	leftDelim = *flagLeftDelim
	rightDelim = *flagRightDelim
	ignore = *flagIgnore

	wasBlank := len(input) == 0

	input = filepath.Clean(input)

	if wasBlank && input == "." {
		panic("** No Directory specified with -i option")
	}

	if len(leftDelim) == 0 {
		panic("** No Left Delimiter specified with -ld option")
	}

	if len(rightDelim) == 0 {
		panic("** No Right Delimiter specified with -rd option")
	}

	if len(ignore) > 0 {
		var err error

		ignoreRegexp, err = regexp.Compile(ignore)
		if err != nil {
			panic("**Error Compiling Regex:" + err.Error())
		}
	}
}
