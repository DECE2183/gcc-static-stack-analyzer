package main

import (
  "fmt"
  "os"
  // "io"
  // "io/fs"
  "regexp"
  "strings"
  "sort"
  // "time"
  // "strconv"
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

type StackInfo struct {
  calls []StackCall
  totalMemUsage int
}

const (
  suFileRegexp = `[.]su$`
  suFileNumbers = `[0-9]+[: ]*?`
  suFileLine = `[a-zA-Z0-9:\s./_]+[\n]`

  ciFileRegexp = `[.]ci$`
)

var (
  stackInfo StackInfo
)

func stringUnspace(s string) string {
  re, _ := regexp.Compile(`\S+`) //(`[a-zA-Z0-9:_/.-=%+]+`)
  strs := re.FindAllString(s, -1)
  return strings.Join(strs, " ")
}

func stringUndrive(s string) string {
  re, _ := regexp.Compile(`[a-zA-Z]:[/\\]`)
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
  _ = regexp.MustCompile(ciFileRegexp)

  if (len(os.Args) < 2) {
    fmt.Println("You must provide a project path as an argument.")
    os.Exit(22)
  }

  projPath := os.Args[1]
  // projFS := os.DirFS(projPath)
  fmt.Printf("Project to analyze: %s\r\n", projPath)

  // Find and analyze all .su files
  var suFiles []string
  EachFile(projPath, suFileRegexp, func(path, content string) {
    suFiles = append(suFiles, path)
    parseFile(&stackInfo.calls, content)
  })

  sort.SliceStable(stackInfo.calls, func(i, j int) bool {
    return stackInfo.calls[i].memUsage > stackInfo.calls[j].memUsage
  })

  stackInfo.totalMemUsage = 0
  for _, call := range stackInfo.calls {
    // fmt.Printf("% 5d: % 8d B %s->%s\r\n", i, call.memUsage, call.fileName, call.entryName)
    stackInfo.totalMemUsage += call.memUsage
  }
  fmt.Printf("Total stack usage: %d Bytes\r\n", stackInfo.totalMemUsage)

  for i := range stackInfo.calls {
    stackInfo.calls[i].memUsagePercent = (float32(stackInfo.calls[i].memUsage) / float32(stackInfo.totalMemUsage))
  }

  if (len(suFiles) < 1) {
    fmt.Println("There are no .su files.")
    os.Exit(22)
  }


  // Find all .ci files
  var ciFiles []string
  var ciGraphs []CodeGraphNode
  EachFile(projPath, ciFileRegexp, func(path, content string) {
    ciFiles = append(ciFiles, path)
    graphs, err := ParseCiFile(content)
    if err != nil {
      fmt.Println(err)
      return
    }
    ciGraphs = append(ciGraphs, graphs...)
  })


  // for _, f_name := range suFiles {
  //   fmt.Println(f_name)
  // }

  // Disable cursor
  // fmt.Println("\e[?25l")
  // defer fmt.Println("\e[?25h")

  // startGUI(&stackInfo)
}
