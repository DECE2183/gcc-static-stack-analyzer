package main

import (
  "os"
  "fmt"
  "strings"
  "strconv"
  "regexp"
)

const (
  windowsAbsolutePathRegexp = "[a-zA-Z]:/"
)

type ParseError struct {
  What string
}

type CodeGraphNode struct {
  Line int
  Column int
  SelfStackUsage int
  MaxStackUsage int
  NodeName string
  FileName string
  EntryName string
  FullSourceFilePath string
  CodeBlock string
  Qualifiers []string
  ChildNodes []*CodeGraphNode
}

func (e *ParseError) Error() string {
  return e.What
}

func getCodeBlock(filePath string, line int) string {
  fileContent, fileError := os.ReadFile(filePath)
  if fileError != nil {
    return fmt.Sprintf("Unable to open source file \"%s\".\n", filePath)
  }
  content := string(fileContent)

  var endlpos int

  // Find proper line
  for li := 0;  li < line - 1; li++ {
    endlpos = strings.Index(content, "\n") + 1
    if endlpos < 1 {
      return fmt.Sprintf("There is no line #%d in source file \"%s\".\n", line, filePath)
    }
    content = content[endlpos:]
  }

  // Read code
  bracesDepthIndex := 0
  endlpos = strings.Index(content, "\n") + 1
  if endlpos < 1 {
    return fmt.Sprintf("Unable to find function start in source file \"%s\".\n", filePath)
  }
  codeString := content[:endlpos]
  content = content[endlpos:]

  if !(strings.Contains(codeString, ";") || strings.Contains(codeString, "}")) {
    // Read all block content if it is not one line
    bracesDepthIndex += strings.Count(codeString, "{")
    for _, line := range strings.Split(content, "\n") {
      bracesDepthIndex += strings.Count(line, "{")
      codeString += line + "\n"
      bracesDepthIndex -= strings.Count(line, "}")
      if bracesDepthIndex <= 0 {
        break
      }
    }
  }

  return codeString
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

  sourceStart := strings.Index(line, "sourcename: \"") + 13
  sourceEnd := sourceStart + strings.Index(line[sourceStart:], "\"")
  sourceName := line[sourceStart:sourceEnd]

  targetStart := strings.Index(line, "targetname: \"") + 13
  targetEnd := targetStart + strings.Index(line[targetStart:], "\"")
  targetName := line[targetStart:targetEnd]

  for _, nodePtr := range baseNode.ChildNodes {
    if nodePtr.NodeName == sourceName {
      sourceNode = nodePtr
    }
    if nodePtr.NodeName == targetName {
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
      winAbsPath, _ := regexp.MatchString(windowsAbsolutePathRegexp, node.FileName)
      if node.FileName[:1] == "/" || winAbsPath {
        node.FullSourceFilePath = node.FileName
      } else {
        node.FullSourceFilePath = baseNode.FullSourceFilePath + "/" + node.FileName
      }
      node.CodeBlock = getCodeBlock(node.FullSourceFilePath, node.Line)
      baseNode.ChildNodes = append(baseNode.ChildNodes, &node)
    } else if strings.Contains(line, "edge: {") {
      baseNode.parseEdge(line)
    }
  }
}

func ParseCiFile(path string) (CodeGraphNode, error) {
  var baseNode CodeGraphNode

  fileContent, fileError := os.ReadFile(path)
  if fileError != nil {
    return baseNode, &ParseError{"File read error."}
  }
  content := string(fileContent)
  baseNode.FullSourceFilePath = strings.ReplaceAll(path, "\\", "/")
  baseNode.FullSourceFilePath = baseNode.FullSourceFilePath[:strings.LastIndex(baseNode.FullSourceFilePath, "/")]
  // baseNode.FullSourceFilePath = baseNode.FullSourceFilePath[:len(baseNode.FullSourceFilePath) - 2]
  // if baseNode.FullSourceFilePath[len(baseNode.FullSourceFilePath) - 1:] == "/" {
  // }

  graphStart := strings.Index(content, "graph: { ")

  for graphStart > -1 {
    graphNameStart := graphStart + strings.Index(content[graphStart:], "\"") + 1
    graphNameEnd := graphNameStart + strings.Index(content[graphNameStart:], "\"")
    graphName := content[graphNameStart:graphNameEnd]

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
  nodePtr.MaxStackUsage = nodePtr.SelfStackUsage
  maxMem := 0

  for _, childNodePtr := range nodePtr.ChildNodes {
    if childNodePtr != nodePtr {
      nodeMaxMem := childNodePtr.CalcStackUsage()
      if nodeMaxMem > maxMem {
        maxMem = nodeMaxMem
      }
    }
  }

  nodePtr.MaxStackUsage = nodePtr.SelfStackUsage + maxMem
  return nodePtr.MaxStackUsage
}
