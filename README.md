# IP2Loc
[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/tserkov/ip2location)
[![Build Status](https://img.shields.io/travis/tserkov/ip2location.svg?style=flat)](https://travis-ci.org/tserkov/ip2location)
[![Coverage Status](https://img.shields.io/coveralls/github/tserkov/ip2location.svg?style=flat)](https://coveralls.io/github/tserkov/ip2location?branch=master)
[![MIT License](https://img.shields.io/github/license/tserkov/ip2location.svg?style=flat)](../master/README.md)

An IP2Location binary data query engine that only works with the DB1 (Country) database.

There are a few major differences between this and the official library, [ip2location/ip2location-go](https://github.com/ip2location/ip2location-go):
- It only supports the DB1 (Country) database;
- Every function can return an error;
- No print statements;
- It's faster and consumes less memory.

## Installation
```
go get -u github.com/tserkov/ip2loc
```

## Usage
It's all in the [godoc](https://godoc.org/github.com/tserkov/ip2location).

### Example
```go
import "github.com/tserkov/ip2loc"
```
```go
db, err := ip2loc.New("/path/to/db.bin")
if err != nil {
	// handle it
}
defer db.Close()

result, err := db.Query("8.8.8.8")
if err != nil {
	// handle it
}

fmt.Printf(
	"Country Code: %s\nCountry Name: %s\n",
	result.CountryCode,
	result.CountryName,
)
```

## Bonus: A binary!
If you happen to be on Linux, this repository contains an executable that accepts space-delimited IPs via arguments or stdin and spits out results in `{IP} {Country Name} ({Country Code})` format to stdout.

It does require that the environment variable `IP2LOC_DB` be set to the path to the ip2location db1 bin.

## License
The super permissive MIT License.

## Contributing
While I don't plan on supporting databases beyond DB1, I will gladly take any performance-enhancing contributions.