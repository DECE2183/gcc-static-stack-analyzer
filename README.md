# GCC static stack analyzer
Terminal UI static stack analysis tool for GCC compiler written in Go.

## Building
```Bash
go mod tidy
go build
```

## Usage
First you need to build the project you want to analyze with `-fcallgraph-info=su` GCC flag. It will generate `.ci` files with function call graph.<br>

Then run the command:
```Bash
./gcc-ssa <path to project build output>
```
