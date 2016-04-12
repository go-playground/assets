package bundler

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProcessedFile contains the information of the processed files.
type ProcessedFile struct {
	OriginalFilename string
	NewFilename      string
}

// BundleDir bundles an entire directory recursively and returns an array of filenames and if an error occurred processing
// suffix will be appended to filenames, if blank a hash of file contents will be added
func BundleDir(dirname string, suffix string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string, ignoreRegexp *regexp.Regexp) ([]*ProcessedFile, error) {
	return bundleDir(dirname, "", false, "", ignoreRegexp, suffix, relativeToDir, relativeDir, leftDelim, rightDelim)
}

func bundleDir(path string, dir string, isSymlinkDir bool, symlinkDir string, ignoreRegexp *regexp.Regexp, output string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) ([]*ProcessedFile, error) {

	var p string
	var processed []*ProcessedFile

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

		if ignoreRegexp != nil && ignoreRegexp.MatchString(fPath) {
			continue
		}

		if file.IsDir() {

			processedFiles, err := bundleDir(p, p, isSymlinkDir, symlinkDir+string(os.PathSeparator)+info.Name(), ignoreRegexp, output, relativeToDir, relativeDir, leftDelim, rightDelim)
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

				processedFiles, err := bundleDir(link, link, true, fPath, ignoreRegexp, output, relativeToDir, relativeDir, leftDelim, rightDelim)
				if err != nil {
					return nil, err
				}

				processed = append(processed, processedFiles...)

				continue
			}
		}

		// process file
		file, err := bundleFile(p, output, relativeToDir, relativeDir, leftDelim, rightDelim, true)
		if err != nil {
			return nil, err
		}

		processed = append(processed, file)
	}

	return processed, nil
}

// BundleFile bundles a single file on disk and returns the filename and if an error occurred processing
func BundleFile(path string, output string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) (*ProcessedFile, error) {
	return bundleFile(path, output, relativeToDir, relativeDir, leftDelim, rightDelim, false)
}

func bundleFile(path string, output string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string, isDirMode bool) (*ProcessedFile, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// fmt.Println("Writing Temp File for:", f.Name())
	newFile, err := ioutil.TempFile("", filepath.Base(f.Name()))
	if err != nil {
		return nil, err
	}

	if err = Bundle(f, newFile, filepath.Dir(path), relativeToDir, relativeDir, leftDelim, rightDelim); err != nil {
		return nil, err
	}

	var newName string

	dirname, filename := filepath.Split(path)
	ext := filepath.Ext(filename)
	filename = filepath.Base(filename)

	if isDirMode && output != "" {
		dirname = output + string(filepath.Separator) + dirname + string(filepath.Separator)
	}

	newName = dirname + filename[0:strings.LastIndex(filename, ext)]

	if isDirMode || output == "" {
		b, err := ioutil.ReadFile(newFile.Name())
		if err != nil {
			return nil, err
		}

		h := md5.New()
		h.Write(b)
		hash := string(h.Sum(nil))

		newName += "-" + fmt.Sprintf("%x", hash) + ext

	} else {
		newName = dirname + output
	}

	abs, err := filepath.Abs(filepath.Dir(newName))
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(abs, os.FileMode(0777)); err != nil {
		return nil, err
	}

	if err = os.Rename(newFile.Name(), newName); err != nil {
		return nil, err
	}

	return &ProcessedFile{OriginalFilename: path, NewFilename: newName}, nil
}

// Bundle combines the given input and writes it out to the provided writer
// removing delims from the combined files
func Bundle(r io.Reader, w io.Writer, dir string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) error {
	return bundle(r, w, dir, relativeToDir, relativeDir, leftDelim, rightDelim, false)
}

// BundleKeepDelims combines the given input and writes it out to the provided writer
// but unlike Bundle() keeps the delims in the combined data
func BundleKeepDelims(r io.Reader, w io.Writer, dir string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string) error {
	return bundle(r, w, dir, relativeToDir, relativeDir, leftDelim, rightDelim, true)
}

func bundle(r io.Reader, w io.Writer, dir string, relativeToDir bool, relativeDir string, leftDelim string, rightDelim string, keepDelims bool) error {

	var err error
	var path string
	var finalPath string

	if !filepath.IsAbs(dir) {
		if dir, err = filepath.Abs(dir); err != nil {
			return err
		}
	}

	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}

	if !fi.IsDir() {

		// check if symlink

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {

			link, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return errors.New("Error Resolving Symlink:" + err.Error())
			}

			fi, err = os.Stat(link)
			if err != nil {
				return err
			}

			if !fi.IsDir() {
				return errors.New("dir passed is not a directory")
			}

			dir = link

		} else {
			return errors.New("dir passed is not a directory")
		}
	}

	l, err := NewLexer("bundle", r, leftDelim, rightDelim)
	if err != nil {
		return err
	}

LOOP:
	for {
		itm := l.NextItem()

		switch itm.Type {
		case ItemLeftDelim, ItemRightDelim:
			if keepDelims {
				w.Write([]byte(itm.Val))
			}
		case ItemText:
			w.Write([]byte(itm.Val))
		case ItemFile:
			if relativeToDir {
				finalPath = dir
				path = relativeDir + "/" + itm.Val
			} else {
				finalPath = filepath.Dir(path)
				path = dir + "/" + itm.Val
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if err = bundle(file, w, finalPath, relativeToDir, relativeDir, leftDelim, rightDelim, keepDelims); err != nil {
				return err
			}
		case ItemEOF:
			break LOOP
		case ItemError:
			return errors.New(itm.Val)
		}
	}

	return nil
}
