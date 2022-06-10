package main

import (
  "fmt"
  "strings"
  "strconv"
)

type ParseError struct {
  What string
}

type CodeGraphNode struct {
  Line int
  Column int
  SelfStackUsage int
  TotalStackUsage int
  NodeName string
  FileName string
  EntryName string
  Qualifiers []string
  ChildNodes []*CodeGraphNode
}

func (e *ParseError) Error() string {
  return e.What
}

func parseNode(line string) CodeGraphNode {
  var newNode CodeGraphNode

  titleStart := strings.Index(line, "title: \"") + 8
  titleEnd := titleStart + strings.Index(line[titleStart:], "\"")
  title := line[titleStart:titleEnd]

  labelStart := strings.Index(line, "label: \"") + 8
  labelEnd := labelStart + strings.Index(line[labelStart:], "\"")
  label := line[labelStart:labelEnd]

  decorators := strings.Split(label, "\\n")

  // Try to get stack usage
  if len(decorators) > 2 {
    if strings.Contains(decorators[2], "bytes") {
      newNode.SelfStackUsage, _ = strconv.Atoi(decorators[2][:strings.Index(decorators[2], " ")])
      qualifiersStart := strings.Index(decorators[2], "(") + 1
      if qualifiersStart > 0 {
        qualifiersEnd := qualifiersStart + strings.Index(decorators[2][qualifiersStart:], ")")
        newNode.Qualifiers = strings.Split(decorators[2][qualifiersStart:qualifiersEnd], ",")
      }
    } else {
      decorators[1] += "\\n" + decorators[2]
      decorators = decorators[:2]
    }
  }

  newNode.NodeName = title
  newNode.EntryName = decorators[0]
  newNode.FileName = strings.ReplaceAll(decorators[1], "\\", "/")

  // Find column index
  numIndex := strings.LastIndex(newNode.FileName, ":")
  newNode.Column, _ = strconv.Atoi(newNode.FileName[numIndex + 1:])
  if numIndex < 0 {
    return newNode
  }
  newNode.FileName = newNode.FileName[:numIndex]

  // Find line index
  numIndex = strings.LastIndex(newNode.FileName, ":")
  newNode.Line, _ = strconv.Atoi(newNode.FileName[numIndex + 1:])
  if numIndex < 0 {
    return newNode
  }
  newNode.FileName = newNode.FileName[:numIndex]

  return newNode
}

func (baseNode *CodeGraphNode) parseEdge(line string) {
  var sourceNode, targetNode *CodeGraphNode

  sourceStart := strings.Index(line, "sourcename: \"") + 8
  sourceEnd := sourceStart + strings.Index(line[sourceStart:], "\"")
  sourcename := line[sourceStart:sourceEnd]

  targetStart := strings.Index(line, "targetname: \"") + 8
  targetEnd := targetStart + strings.Index(line[targetStart:], "\"")
  targetname := line[targetStart:targetEnd]

  for _, nodePtr := range baseNode.ChildNodes {
    if nodePtr.NodeName == sourcename {
      sourceNode = nodePtr
    }
    if nodePtr.NodeName == targetname {
      targetNode = nodePtr
    }
  }

  if sourceNode == nil || targetNode == nil {
    return
  }

  if sourceNode == targetNode {
    sourceNode.Qualifiers = append(sourceNode.Qualifiers, "recursive")
  }

  sourceNode.ChildNodes = append(sourceNode.ChildNodes, targetNode)
}

func (baseNode *CodeGraphNode) parseGraph(name, content string) {
  lines := strings.Split(content, "\r\n")

  for _, line := range lines {
    if strings.Contains(line, "node: {") {
      node := parseNode(line)
      baseNode.ChildNodes = append(baseNode.ChildNodes, &node)
      // fmt.Printf("Node: %v\n", node)
    } else if strings.Contains(line, "edge: {") {
      baseNode.parseEdge(line)
    }
  }
}

func ParseCiFile(content string) (CodeGraphNode, error) {
  var baseNode CodeGraphNode
  graphStart := strings.Index(content, "graph: { ")

  for graphStart > -1 {
    graphNameStart := graphStart + strings.Index(content[graphStart:], "\"") + 1
    graphNameEnd := graphNameStart + strings.Index(content[graphNameStart:], "\"")
    graphName := content[graphNameStart:graphNameEnd]

    fmt.Printf("Found graph: \"%s\"\n", graphName)

    graphStart = graphStart + strings.Index(content[graphStart:], "\n") + 1
    graphEnd := graphStart + strings.LastIndex(content[graphStart:], "}\r\n")
    baseNode.parseGraph(graphName, content[graphStart:graphEnd])

    content = content[graphEnd+1:]
    graphStart = strings.Index(content, "graph: { ")
  }

  if (len(baseNode.ChildNodes) > 0) {
    return baseNode, nil
  } else {
    return baseNode, &ParseError{"Graphs not found."}
  }
}

func (nodePtr *CodeGraphNode) CalcStackUsage() int {
  nodePtr.TotalStackUsage = nodePtr.SelfStackUsage
  for _, childNodePtr := range nodePtr.ChildNodes {
    if childNodePtr == nodePtr {
      continue
    }
    nodePtr.TotalStackUsage += childNodePtr.CalcStackUsage()
  }
  return nodePtr.TotalStackUsage
}
