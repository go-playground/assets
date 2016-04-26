package assets

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/bundler"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

const (
	oldNewSeparator = " --> "
	manifestFile    = "/manifest.txt"
	cssHTMLTag      = "css_tag"
	jsHTMLTag       = "js_tag"
)

// RunMode is the type that determines which mode the template.FuncMap functions whould run in.
type RunMode int

// RunMode's
const (
	Development RunMode = iota
	Production
)

const (
	jsTag  = `<script type="text/javascript" src="%s"></script>`
	cssTag = `<link type="text/css" rel="stylesheet" href="%s">`
)

var m *minify.M

func initMinifier() {

	if m != nil {
		return
	}

	m = minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
}

// Generate processes (bundles, compresses...) the assets for use and creates the Manifest file
// NOTE: no compression yet until there is a native and establishes compressor written in Go
func Generate(dirname string, outputDir string, relativeToDir bool, leftDelim string, rightDelim string, extensions map[string]struct{}) ([]*bundler.ProcessedFile, string, error) {

	initMinifier()

	dirname = filepath.Clean(dirname)
	outputDir = filepath.Clean(outputDir) + string(filepath.Separator)

	abs, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, "", err
	}

	if err = os.MkdirAll(abs, os.FileMode(0777)); err != nil {
		return nil, "", err
	}

	// verify dirname is actually a DIR + do symlink check
	fi, err := os.Lstat(dirname)
	if err != nil {
		return nil, "", err
	}

	if !fi.IsDir() {

		// check if symlink
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {

			link, err := filepath.EvalSymlinks(dirname)
			if err != nil {
				return nil, "", errors.New("Error Resolving Symlink:" + err.Error())
			}

			fi, err = os.Stat(link)
			if err != nil {
				return nil, "", err
			}

			if !fi.IsDir() {
				return nil, "", errors.New("dirname passed in is not a directory")
			}

			dirname = link

		} else {
			return nil, "", errors.New("dirname passed in is not a directory")
		}
	}

	var manifest string

	if outputDir == "" {
		manifest = dirname + manifestFile
	} else {
		manifest = outputDir + dirname + manifestFile
	}

	f, err := os.Open(manifest)
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			files := strings.SplitN(scanner.Text(), oldNewSeparator, 2)

			fmt.Println("Removing Existing File:", files[1])
			os.Remove(files[1])
		}

		os.Remove(manifest)
	}

	processed, err := bundleDir(dirname, "", false, "", extensions, outputDir, relativeToDir, dirname, leftDelim, rightDelim)
	if err != nil {
		return nil, "", err
	}

	var buff bytes.Buffer

	for _, file := range processed {
		buff.WriteString(filepath.FromSlash(file.OriginalFilename))
		buff.WriteString(oldNewSeparator)
		buff.WriteString(filepath.FromSlash(file.NewFilename))
		buff.WriteString("\n")
	}

	if err = ioutil.WriteFile(manifest, buff.Bytes(), 0644); err != nil {
		return nil, "", err
	}

	return processed, manifest, nil
}

func bundleDir(path string, dir string, isSymlinkDir bool, symlinkDir string, extensions map[string]struct{}, output string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) ([]*bundler.ProcessedFile, error) {

	var p string
	var ext string
	var processed []*bundler.ProcessedFile

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, file := range files {

		info := file
		p = path + string(os.PathSeparator) + file.Name()
		fPath := p

		if isSymlinkDir {
			fPath = strings.Replace(p, dir, symlinkDir, 1)
		}

		if file.IsDir() {

			processedFiles, err := bundleDir(p, p, isSymlinkDir, symlinkDir+string(os.PathSeparator)+info.Name(), extensions, output, relativeToDir, relativeDir, leftDelim, rightDelim)
			if err != nil {
				return nil, err
			}

			processed = append(processed, processedFiles...)

			continue
		}

		if file.Mode()&os.ModeSymlink == os.ModeSymlink {

			link, err := filepath.EvalSymlinks(p)
			if err != nil {
				log.Panic("Error Resolving Symlink", err)
			}

			fi, err := os.Stat(link)
			if err != nil {
				log.Panic(err)
			}

			info = fi

			if fi.IsDir() {

				processedFiles, err := bundleDir(link, link, true, fPath, extensions, output, relativeToDir, relativeDir, leftDelim, rightDelim)
				if err != nil {
					return nil, err
				}

				processed = append(processed, processedFiles...)

				continue
			}
		}

		// if we get here, it's a file
		ext = filepath.Ext(fPath)

		if _, ok := extensions[ext]; !ok {

			// just copy file to final location
			if err := copyFile(fPath, output); err != nil {
				return nil, err
			}
			continue
		}

		// process file
		file, err := bundleFile(p, output, relativeToDir, relativeDir, leftDelim, rightDelim, ext)
		if err != nil {
			return nil, err
		}

		processed = append(processed, file)
	}

	return processed, nil
}

func copyFile(path string, output string) error {

	dirname := output + string(filepath.Separator) + path

	if err := os.MkdirAll(filepath.Dir(dirname), os.FileMode(0777)); err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := os.Create(dirname)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, f)
	if err != nil {
		return err
	}

	return nil
}

func bundleFile(path string, output string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string, extension string) (*bundler.ProcessedFile, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	newFile, err := ioutil.TempFile("", filepath.Base(f.Name()))
	if err != nil {
		return nil, err
	}

	defer newFile.Close()

	if err = bundler.Bundle(f, newFile, filepath.Dir(path), relativeToDir, relativeDir, leftDelim, rightDelim); err != nil {
		return nil, err
	}

	var newName string

	dirname, filename := filepath.Split(path)
	origDir := dirname
	ext := filepath.Ext(filename)
	filename = filepath.Base(filename)

	if output != "" {
		dirname = output + string(filepath.Separator) + dirname + string(filepath.Separator)
	}

	newName = dirname + filename[0:strings.LastIndex(filename, ext)]
	origDir += filename[0:strings.LastIndex(filename, ext)]

	b, err := ioutil.ReadFile(newFile.Name())
	if err != nil {
		return nil, err
	}

	h := md5.New()
	h.Write(b)
	hash := string(h.Sum(nil))

	hashName := "-" + fmt.Sprintf("%x", hash) + ext
	newName += hashName
	origDir += hashName

	if err = os.MkdirAll(filepath.Dir(newName), os.FileMode(0777)); err != nil {
		return nil, err
	}

	buff := new(bytes.Buffer)

	// perform minification
	if extension == ".js" {

		if err := m.Minify("text/javascript", buff, bytes.NewReader(b)); err != nil {
			return nil, err
		}

	} else if extension == ".css" {
		if err := m.Minify("text/css", buff, bytes.NewReader(b)); err != nil {
			return nil, err
		}
	}

	if err := os.Remove(newFile.Name()); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(newName, buff.Bytes(), 0644); err != nil {
		return nil, err
	}

	return &bundler.ProcessedFile{OriginalFilename: path, NewFilename: origDir}, nil
}

// LoadManifestFiles reads the manifest file generated by the Generate() command
// in Production mode and returns template.FuncMap for the provided RunMode
func LoadManifestFiles(dirname string, mode RunMode, relativeToDir bool, leftDelim string, rightDelim string) (template.FuncMap, error) {

	var f *os.File
	var err error

	if mode == Production {
		f, err = os.Open(dirname + manifestFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}

	return ProcessManifestFiles(f, dirname, mode, relativeToDir, leftDelim, rightDelim)
}

// ProcessManifestFiles reads an existing manifest file generated by the Generate() command
// in Production mode and returns template.FuncMap for the provided RunMode
func ProcessManifestFiles(manifest io.Reader, dirname string, mode RunMode, relativeToDir bool, leftDelim string, rightDelim string) (template.FuncMap, error) {

	mapped := map[string]string{}
	dirname = filepath.Clean(dirname) + string(os.PathSeparator)

	if mode == Production {

		var files []string

		scanner := bufio.NewScanner(manifest)

		for scanner.Scan() {

			files = strings.SplitN(scanner.Text(), oldNewSeparator, 2)
			mapped[strings.TrimLeft(files[0], dirname)] = "/" + files[1]
		}
	}

	return loadMapFuncs(dirname, mode, relativeToDir, leftDelim, rightDelim, mapped), nil
}

func loadMapFuncs(dirname string, mode RunMode, relativeToDir bool, leftDelim string, rightDelim string, mapped map[string]string) template.FuncMap {

	funcs := template.FuncMap{}

	if mode == Production {
		funcs[cssHTMLTag] = createProdCSSTemplateFunc(mapped)
		funcs[jsHTMLTag] = createProdJSTemplateFunc(mapped)

		return funcs
	}

	funcs[cssHTMLTag] = createDevCSSTemplateFunc(dirname, relativeToDir, leftDelim, rightDelim)
	funcs[jsHTMLTag] = createDevJSTemplateFunc(dirname, relativeToDir, leftDelim, rightDelim)

	return funcs
}

func createProdCSSTemplateFunc(mapped map[string]string) interface{} {
	return func(name string) template.HTML {
		return template.HTML(fmt.Sprintf(cssTag, mapped[name]))
	}
}

func createProdJSTemplateFunc(mapped map[string]string) interface{} {
	return func(name string) template.HTML {
		return template.HTML(fmt.Sprintf(jsTag, mapped[name]))
	}
}

func createDevCSSTemplateFunc(dirname string, relativeToDir bool, leftDelim string, rightDelim string) interface{} {
	// custom lexer, bytesBuffer

	return func(name string) template.HTML {
		buff := new(bytes.Buffer)

		files, err := loadFromDelims(dirname, name, relativeToDir, dirname, leftDelim, rightDelim)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			buff.WriteString(fmt.Sprintf(cssTag, "/"+filepath.Clean(dirname)+file))
		}

		buff.WriteString(fmt.Sprintf(cssTag, "/"+dirname+name))

		return template.HTML(buff.String())
	}
}

func createDevJSTemplateFunc(dirname string, relativeToDir bool, leftDelim string, rightDelim string) interface{} {
	// custom lexer, bytesBuffer

	return func(name string) template.HTML {
		buff := new(bytes.Buffer)

		files, err := loadFromDelims(dirname, name, relativeToDir, dirname, leftDelim, rightDelim)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			buff.WriteString(fmt.Sprintf(jsTag, "/"+filepath.Clean(dirname)+file))
		}

		buff.WriteString(fmt.Sprintf(jsTag, "/"+dirname+name))

		return template.HTML(buff.String())
	}
}

func loadFromDelims(dirname string, name string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) ([]string, error) {
	var err error
	var files []string
	var ok bool
	var path string

	existing := map[string]struct{}{}

	f, err := os.Open(dirname + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	l, err := bundler.NewLexer("assets-bundle", f, leftDelim, rightDelim)
	if err != nil {
		return nil, err
	}

LOOP:
	for {
		itm := l.NextItem()

		switch itm.Type {

		case bundler.ItemFile:

			if relativeToDir {
				path = filepath.FromSlash("/" + itm.Val)
			} else {
				path = filepath.FromSlash("/" + dirname + itm.Val)
			}

			if _, ok = existing[path]; !ok {
				files = append(files, path)
				existing[path] = struct{}{}
			}

			fls, err := loadFromDelims(dirname, itm.Val, relativeToDir, relativeDir, leftDelim, rightDelim)
			if err != nil {
				return nil, err
			}

			// must prepend as the just processed files are requirements.
			files = append(fls, files...)

		case bundler.ItemEOF:
			break LOOP
		case bundler.ItemError:
			return nil, errors.New(itm.Val)
		}
	}

	return files, nil
}
