package assets

import (
	"os"
	"testing"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

var extensions = map[string]struct{}{".txt": {}}

func TestGenerate(t *testing.T) {

	// func Generate(dirname string, outputDir string, relativeToDir bool, leftDelim string, rightDelim string, ignoreRegexp *regexp.Regexp) ([]*bundler.ProcessedFile, string, error) {
	processed, manifest, err := Generate("testfiles/test1", "testfiles/test1output", false, "include(", ")", extensions)
	Equal(t, err, nil)
	Equal(t, manifest, "testfiles/test1output/testfiles/test1/manifest.txt")
	Equal(t, len(processed), 3)

	// Equal(t, processed[0].OriginalFilename, "testfiles/test1/file1.txt")
	// Equal(t, processed[1].OriginalFilename, "testfiles/test1/file2.txt")
	// Equal(t, processed[2].OriginalFilename, "testfiles/test1/file3.txt")

	funcs, err := LoadManifestFiles("testfiles/test1output", Development, false, "include(", ")")
	Equal(t, err, nil)
	Equal(t, len(funcs), 2)

	err = os.RemoveAll("testfiles/test1output")
	Equal(t, err, nil)

	// test BAD input
	_, err = LoadManifestFiles("testfiles/test1outputofnonexistant...", Production, false, "include(", ")")
	NotEqual(t, err, nil)
	Equal(t, err.Error(), "open testfiles/test1outputofnonexistant.../manifest.txt: no such file or directory")
}

func TestGenerateWithSymlinks(t *testing.T) {

	// func Generate(dirname string, outputDir string, relativeToDir bool, leftDelim string, rightDelim string, ignoreRegexp *regexp.Regexp) ([]*bundler.ProcessedFile, string, error) {
	processed, manifest, err := Generate("testfiles/test2", "testfiles/test2output", false, "include(", ")", extensions)
	Equal(t, err, nil)
	Equal(t, manifest, "testfiles/test2output/testfiles/test2/manifest.txt")
	Equal(t, len(processed), 5)

	// Equal(t, processed[0].OriginalFilename, "testfiles/test2/f2.txt")
	// Equal(t, processed[1].OriginalFilename, "testfiles/test2/file1.txt")
	// Equal(t, processed[2].OriginalFilename, "testfiles/test2/inner/file2.txt")

	funcs, err := LoadManifestFiles("testfiles/test2output", Development, false, "include(", ")")
	Equal(t, err, nil)
	Equal(t, len(funcs), 2)

	err = os.RemoveAll("testfiles/test2output")
	Equal(t, err, nil)
}
