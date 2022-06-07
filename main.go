package main

import (
  "fmt"
  "os"
  // "io"
  "io/fs"
  "regexp"
  "strings"
  "sort"
  "time"
  // "strconv"
)

const (
  suFileRegexp = `[.]su$`
  suFileNumbers = `[0-9]+[: ]*?`
  suFileLine = `[a-zA-Z0-9:\s./_]+[\n]`
)

type StackCall struct {
  line int
  column int
  memUsage int
  memUsagePercent float32
  fileName string
  entryName string
  qualifiers string
}

func stringUnspace(s string) string {
  re, _ := regexp.Compile(`\S+`) //(`[a-zA-Z0-9:_/.-=%+]+`)
  strs := re.FindAllString(s, -1)
  return strings.Join(strs, " ")
}

func stringUndrive(s string) string {
  re, _ := regexp.Compile(`[A-Z]:[/\\]`)
  if (re.MatchString(s)) {
    return s[2:]
  }
  return s
}

func parseFile(calls *[]StackCall, str string) {
  lines := strings.Split(strings.ReplaceAll(str, "\r\n", "\n"), "\n")
  for _, line := range lines {
    if (len(line) < 2) {
      continue
    }

    var newCall StackCall;

    line = stringUnspace(line)
    line = stringUndrive(line)
    line = strings.ReplaceAll(line, ":", " ")

    fmt.Sscanf(line, "%s %d %d %s %d %s",
      &newCall.fileName,
      &newCall.line,
      &newCall.column,
      &newCall.entryName,
      &newCall.memUsage,
      &newCall.qualifiers)

    *calls = append(*calls, newCall)
  }
}

func main() {
  _ = regexp.MustCompile(suFileRegexp)
  _ = regexp.MustCompile(suFileNumbers)
  _ = regexp.MustCompile(suFileLine)

  if (len(os.Args) < 2) {
    fmt.Println("You must provide a project path as an argument.")
    os.Exit(22)
  }

  projPath := os.Args[1]
  projFS := os.DirFS(projPath)
  fmt.Printf("Project to analyze: %s\r\n", projPath)

  // Find and analyze all .su files
  var suFiles []string;
  var suCalls []StackCall;
  fs.WalkDir(projFS, ".", func (path string, dir fs.DirEntry, err error) error {
    if (err != nil) {
      return fs.SkipDir
    }
    match, _ := regexp.MatchString(suFileRegexp, path)
    if (match) {
      suFiles = append(suFiles, path)
      fileContent, fError := os.ReadFile(fmt.Sprintf("%s/%s", projPath, path))
      if (fError != nil) {
        return fError
      }
      fileString := string(fileContent)
      parseFile(&suCalls, fileString)
    }
    return nil
  })

  sort.SliceStable(suCalls, func(i, j int) bool {
    return suCalls[i].memUsage > suCalls[j].memUsage
  })

  totalUsage := 0
  for i, call := range suCalls {
    fmt.Printf("% 5d: % 8d B %s->%s\r\n", i, call.memUsage, call.fileName, call.entryName)
    totalUsage += call.memUsage
  }
  fmt.Printf("Total stack usage: %d Bytes\r\n", totalUsage)

  for i := range suCalls {
    suCalls[i].memUsagePercent = (float32(suCalls[i].memUsage) / float32(totalUsage))
  }

  if (len(suFiles) < 1) {
    fmt.Println("There are no .su files.")
    os.Exit(22)
  }

  // for _, f_name := range suFiles {
  //   fmt.Println(f_name)
  // }

  // Disable cursor
  // fmt.Println("\e[?25l")
  // defer fmt.Println("\e[?25h")

  for {
  	drawGUI(suCalls, totalUsage)
  	time.Sleep(26 * time.Millisecond)
  }
}
