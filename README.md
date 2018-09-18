[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jbsmith7741/uri)
[![Build Status](https://travis-ci.com/jbsmith7741/uri.svg?branch=master)](https://travis-ci.com/jbsmith7741/uri)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/uri)](https://goreportcard.com/report/github.com/jbsmith7741/uri)
[![codecov](https://codecov.io/gh/jbsmith7741/uri/branch/master/graph/badge.svg)](https://codecov.io/gh/jbsmith7741/uri)

# uri
a convenient and easy way to unmarshal a uri to a struct.
 
## keywords
- schema
- host
- path
- authority (schema:host)
- origin (schema:host/path)


## example
If we have the uri "http://example.com/path/to/page?name=ferret&color=purple" we can unmarshal this to a predefined struct as follows
``` go 
type Example struct {
    Schema `uri:"schema"`
    Host   `uri:"Host"`
    Path   `uri:"path"`
    Name   `uri:"name"`
    Color  `uri:"color"`
}

func() {
e := Example{}

err := uri.Unmarshal("http://example.com/path/to/page?name=ferret&color=purple", &e)
 
}
```
this would become the following struct 
``` go
e := Example{
    Schema: "www",
    Host:   "example.com",
    Path:   "path/to/page",
    Name:   "ferret",
    Color:  "purple",
    }
 
```

## example 

``` golang 
uri = http://example.org/wiki/Main_Page?Option1=10&Text=hello 

type MyStruct struct {
    Schema `uri:"scheme"`
    Host `uri:"host"`
    Path `uri:"path"`
    Option1 int
    Text string 
}

func Parse() {
    var s *MyStruct
    uri.Unmarshal(s, uri)
}
```
