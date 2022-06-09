package main

import (
  "fmt"
  "strings"
)

type ParseError struct {
  What string
}

type CodeGraphNode struct {
  FileName string
  EntryName string
  Line int
  Column int
  ChildNodes []CodeGraphNode
}

func (e *ParseError) Error() string {
  return e.What
}

func parseNode(line string) CodeGraphNode {

}

func (graph *CodeGraphNode) parseEdge(line string) {

}

func parseGraph(name, content string) CodeGraphNode {
  var newGraph CodeGraphNode

  var lines := strings.Split(content, "\r\n")
  for i, line := range lines {
    if strings.Contains(line, "node: {") {
      newGraph.ChildNodes = append(newGraph.ChildNodes, parseNode(line))
    } else if strings.Contains(line, "edge: {") {
      newGraph.parseEdge(line)
    }
  }

  return newGraph
}

func ParseCiFile(content string) ([]CodeGraphNode, error) {
  var graphs []CodeGraphNode
  graphStart := strings.Index(content, "graph: { ")

  for graphStart > -1 {
    graphNameStart := graphStart + strings.Index(content[graphStart:], "\"") + 1
    graphNameEnd := graphNameStart + strings.Index(content[graphNameStart:], "\"")
    graphName := content[graphNameStart:graphNameEnd]

    fmt.Printf("Found graph: \"%s\"\n", graphName)

    graphStart = graphStart + strings.Index(content[graphStart:], "\n") + 1
    graphEnd := graphStart + strings.Index(content[graphStart:], "}\r\n")
    graphs = append(graphs, parseGraph(graphName, content[graphStart:graphEnd]))

    content = content[graphEnd+1:]
    graphStart = strings.Index(content, "graph: { ")
  }

  if (len(graphs) > 0) {
    return graphs, nil
  } else {
    return graphs, &ParseError{"Graphs not found."}
  }
}
