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
	flagFileOrDir             = flag.String("i", "", "Directory to bundle files for recursivly.")
	flagOuputFile             = flag.String("o", "", "If using a DIR in -i option then this will be the output file directory, if blank will use -i option DIR.")
	flagLeftDelim             = flag.String("ld", "", "The Left Delimiter for file includes")
	flagRightDelim            = flag.String("rd", "", "The Right Delimiter for file includes")
	flagIncludesRelativeToDir = flag.Bool("rtd", true, "Specifies if the files included should be treated as relative to the directory, or relative to the files from which they are included.")
	flagIgnore                = flag.String("ignore", "", "Regexp for files/dirs we should ignore i.e. \\.gitignore; only used when -i option is a DIR")

	input      string
	output     string
	leftDelim  string
	rightDelim string
	ignore     string

	relativeToDir bool

	ignoreRegexp *regexp.Regexp

	processed []*bundler.ProcessedFile
)

func main() {
	parseFlags()

	processed, manifest, err := assets.Generate(input, output, relativeToDir, leftDelim, rightDelim, ignoreRegexp)
	if err != nil {
		panic(err)
	}

	printResults(processed)

	fmt.Println("\nManifest Generated:", manifest)
	fmt.Printf("\n")
}

func printResults(processed []*bundler.ProcessedFile) {

	fmt.Printf("The following files were processed:\n\n")

	for _, file := range processed {
		fmt.Println("  " + file.NewFilename)
	}
}

func parseFlags() {

	flag.Parse()

	input = strings.TrimSpace(*flagFileOrDir)
	output = strings.TrimSpace(*flagOuputFile)
	leftDelim = *flagLeftDelim
	rightDelim = *flagRightDelim
	relativeToDir = *flagIncludesRelativeToDir
	ignore = *flagIgnore

	wasBlank := len(input) == 0

	input = filepath.Clean(input)
	output = filepath.Clean(output)

	if wasBlank && input == "." {
		panic("** No Directory specified with -i option")
	}

	if len(output) == 0 {
		output = input
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
