[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jbsmith7741/uri)
[![Build Status](https://travis-ci.com/jbsmith7741/uri.svg?branch=master)](https://travis-ci.com/jbsmith7741/uri)
[![Go Report Card](https://goreportcard.com/badge/github.com/jbsmith7741/uri)](https://goreportcard.com/report/github.com/jbsmith7741/uri)
[![codecov](https://codecov.io/gh/jbsmith7741/uri/branch/master/graph/badge.svg)](https://codecov.io/gh/jbsmith7741/uri)

Support for go 1.10+ 

Older versions will probably work, but are not officially supported or tested against 

# uri

a convenient and easy way to convert from a uri to a struct or vic-versa

## keywords

- scheme
- host
- path
- authority (schema:host)
- origin (schema:host/path)

## struct tags

- **uri** - the name of the variable or to designate a special keywords (schema, host, etc). empty defaults the exact name of the struct (same as json tags)
- **default** - defined the default value of a variable
- **required** - if the param is missing, unmarshal will return an error
- **format** - time format field for marshaling of time.Time


## example 1

If we have the uri "http://example.com/path/to/page?name=ferret&color=purple" we can unmarshal this to a predefined struct as follows

``` go
type Example struct {
    Scheme string `uri:"scheme"`
    Host   string `uri:"Host"`
    Path   string `uri:"path"`
    Name   string `uri:"name"`
    Color  string `uri:"color"`
}

func() {
e := Example{}

err := uri.Unmarshal("http://example.com/path/to/page?name=ferret&color=purple", &e)
}
```

this would become the following struct

``` go
e := Example{
    Schema: "http",
    Host:   "example.com",
    Path:   "path/to/page",
    Name:   "ferret",
    Color:  "purple",
    }
```

## example 2 - defaults

``` go
var site = "http://example.org/wiki/Main_Page?Option1=10"

type MyStruct struct {
    Path    string `uri:"path"`
    Option1 int
    Text    string `default:"qwerty"`
}

func Parse() {
    s := &MyStruct{}
    uri.Unmarshal(site, s)
}
```

this becomes

``` go
e := &MyStruct{
    Path: "/wiki/Main_Page"
    Option1: 10,
    Text: "qwerty",
}
```

## example 3 - required field

``` go
type Example struct {
    Name string `uri:"name"`
    Token string `uri:"token" required:"true"`
}
func Parse() {
   site := "?name=hello"
   e := &Example{}
   err := uri.Unmarshal(site, e)
}
```
Result
```
    token is required
```