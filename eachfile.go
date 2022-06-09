package main

import (
  "io/fs"
  "os"
  "regexp"
  "fmt"
)

type EachCallback func(path, content string)

func EachFile(path, pattern string, callback EachCallback) {
  dirFS := os.DirFS(path)

  fs.WalkDir(dirFS, ".", func (_path string, _dir fs.DirEntry, _err error) error {
    if (_err != nil) {
      return fs.SkipDir
    }
    match, _ := regexp.MatchString(pattern, _path)
    if (match) {
      fileContent, fileError := os.ReadFile(fmt.Sprintf("%s/%s", path, _path))
      if (fileError != nil) {
        return fileError
      }
      fileString := string(fileContent)
      callback(_path, fileString)
    }
    return nil
  })
}
