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

  // Find all .ci files
  var ciGraph CodeGraphNode
  EachFile(projPath, ciFileRegexp, func(path string) {
    graph, err := ParseCiFile(path)
    if err != nil {
      fmt.Println(err)
      return
    }
    ciGraph.ChildNodes = append(ciGraph.ChildNodes, graph.ChildNodes...)
  })

  if (len(ciGraph.ChildNodes) < 1) {
    fmt.Println("There are no .ci files. Be sure to add \"-fcallgraph-info=su\" arguments to the build command.")
    os.Exit(22)
  }

  _ = ciGraph.CalcStackUsage()

  sort.SliceStable(ciGraph.ChildNodes, func(i, j int) bool {
    return ciGraph.ChildNodes[i].MaxStackUsage > ciGraph.ChildNodes[j].MaxStackUsage
  })

  startGUI(&ciGraph)
}
