# orbit 

A Node.js lambda environment library for Go (Golang).

[![](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/abcum/orbit) [![](https://goreportcard.com/badge/github.com/abcum/orbit?style=flat-square)](https://goreportcard.com/report/github.com/abcum/orbit) [![](https://img.shields.io/badge/license-Apache_License_2.0-blue.svg?style=flat-square)](https://github.com/abcum/orbit) 

#### Features

- Pre-load npm modules directly
- Pre-load npm modules from files
- Pre-load npm modules from folders 
- Interrupt long-running code
- Configurable stack trace depth limit
- setTimeout, setInterval, setImmediate built-in
- Callback for finding npm modules not just on filesystem
- Easily configure runtime before/after npm modules are loaded
- Track time spent running code in Node.js environment

#### Installation

```
go get github.com/abcum/orbit
```
