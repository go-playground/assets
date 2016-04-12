## bundler

[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/bundler/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/bundler)
[![Go Report Card](http://goreportcard.com/badge/go-playground/bundler)](http://goreportcard.com/badge/go-playground/bundler)
[![GoDoc](https://godoc.org/github.com/go-playground/bundler?status.svg)](https://godoc.org/github.com/go-playground/bundler)

bundler is a generic file combiner of any type of files using a custom left and right delimiter, i.e. css or js files using a lexer. bundler can be used via command line or as a library.

#### Why
--------
I was writing an asset pipeline and needed a way to combine the files using some sort of include format and so why not a flexible library to handle that and can be used elsewhere.

#### Best Part
----------
uses a Lexer to process the files, with necessary components exposed to use as you wish.

#### Installation
------------------
Use go get
```go
go get github.com/go-playground/bundler/cmd/bundler
``` 

or to update

```go
go get -u github.com/go-playground/bundler/cmd/bundler
``` 

Then import lars package into your code.

```go
import "github.com/go-playground/bundler"
```

#### Command Line Use
--------------
```
bundler -h
Usage of ./bundler:
  -i string
    	File or DIR to bundle files for; DIR will bundle all files within the DIR recursivly.
  -ignore string
    	Regexp for files/dirs we should ignore i.e. \.gitignore; only used when -i option is a DIR
  -ld string
    	the Left Delimiter for file includes, if not specified default to //include(. (default "//include(")
  -o string
    	Output filename, or if using a DIR in -i option the suffix, otherwise will be be the filename with appended hash of file contents.
  -rd string
    	the Right Delimiter for file includes, if not specified default to ). (default ")")
  ```
