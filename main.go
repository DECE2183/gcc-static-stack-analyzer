package main

import (
  "fmt"
  "os"
  "sort"
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
  ciFileRegexp = `[.]ci$`
)

var (
  stackInfo StackInfo
)

func main() {
  if (len(os.Args) < 2) {
    fmt.Println("You must provide a project path as an argument.")
    os.Exit(22)
  }

  projPath := os.Args[1]
  fmt.Printf("Project to analyze: %s\r\n", projPath)

  // Find all .ci files
  var ciFiles []string
  var ciGraph CodeGraphNode
  EachFile(projPath, ciFileRegexp, func(path, content string) {
    ciFiles = append(ciFiles, path)
    graph, err := ParseCiFile(content)
    if err != nil {
      fmt.Println(err)
      return
    }
    ciGraph.ChildNodes = append(ciGraph.ChildNodes, graph.ChildNodes...)
  })

  if (len(ciGraph.ChildNodes) < 1) {
    fmt.Println("There are no .su files. Be sure to add \"-fstack-usage -fcallgraph-info=su\" arguments to the build command.")
    os.Exit(22)
  }

  worstStackUsage := ciGraph.CalcStackUsage()

  sort.SliceStable(ciGraph.ChildNodes, func(i, j int) bool {
    return ciGraph.ChildNodes[i].MaxStackUsage > ciGraph.ChildNodes[j].MaxStackUsage
  })

  fmt.Printf("Worst stack usage: %d Bytes\r\n", worstStackUsage)

  startGUI(&ciGraph)
}
