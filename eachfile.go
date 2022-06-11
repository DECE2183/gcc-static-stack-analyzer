package main

import (
  "io/fs"
  "os"
  "fmt"
  "regexp"
)

type EachCallback func(path string)

func EachFile(path, pattern string, callback EachCallback) {
  r, rerr := regexp.Compile(pattern)
  if rerr != nil {
    return
  }
  dirFS := os.DirFS(path)

  fs.WalkDir(dirFS, ".", func (_path string, _dir fs.DirEntry, _err error) error {
    if (_err != nil) {
      return fs.SkipDir
    }
    match := r.MatchString(_path)
    if (match) {
      callback(fmt.Sprintf("%s/%s", path, _path))
    }
    return nil
  })
}
