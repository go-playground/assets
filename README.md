## assets
[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/assets-2/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/assets-2)
[![Go Report Card](http://goreportcard.com/badge/go-playground/assets)](http://goreportcard.com/badge/go-playground/assets)
[![GoDoc](https://godoc.org/github.com/go-playground/assets?status.svg)](https://godoc.org/github.com/go-playground/assets)

assets is a simple + lightweight Asset Pipeline for Go HTML applications

#### Why
--------
Couldn't find a good one that wasn't a mess or didn't rely on some other language to compress the assets.

#### What's missing
----------
Currently this library doesn't compress the combined files, when using Gzip it's not that big of a deal, however 
once an established js and css compressor is created it can be added seamlessly; **pull requests are welcome to help**

#### Installation
------------------
Use go get
```go
go get github.com/go-playground/assets/cmd/assets
``` 

or to update

```go
go get -u github.com/go-playground/assets/cmd/assets
``` 

Then import assets package into your code.

```go
import "github.com/go-playground/assets"
```

#### Command Line Use
--------------
```
assets -h

Usage of assets:
  -i string
    	Asset directory to bundle files for recursivly.
  -ignore string
    	Regexp for files/dirs we should ignore i.e. \.gitignore.
  -ld string
    	The Left Delimiter for file includes
  -o string
    	Output directory, if blank will use -i option DIR.
  -rd string
    	The Right Delimiter for file includes
  -rtd
    	Specifies if the files included should be treated as relative to the directory, or relative to the files from which they are included. (default true)
  ```
