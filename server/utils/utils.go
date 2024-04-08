package utils

import (
	"strings"
	"time"

	"github.com/remiges-tech/alya/wscutils"
)

const (
	DIALTIMEOUT        = 50 * time.Second
	RIGELPREFIX        = "/remiges/rigel"
	INVALID_DEPENDENCY = "invalid_dependency"

	ErrcodeMissingRequiredFields = "missing_required_fields"
)

type Node struct {
	Name     string
	Children map[string]*Node
	IsLeaf   bool
	FullPath string
	Value    string
}

func NewNode(name string) *Node {
	return &Node{
		Name:     name,
		Children: make(map[string]*Node),
		IsLeaf:   false,
		FullPath: "",
		Value:    "",
	}
}

// add nodes corresponding to the path in the node tree
func (n *Node) AddPath(path string, val string) {
	// []parts := split the path on '/'
	var parts []string
	parts = strings.Split(path, "/")
	fullPath := ""

	current := n
	// fmt.Printf("root: %v \n", current.Name)
	for i, part := range parts {
		if i == 0 {
			continue
		}
		_, exists := current.Children[part]
		// fmt.Printf("check exist %v \n", part)
		if !exists {
			current.Children[part] = NewNode(part)
			// fmt.Printf("part miss: %v \n", part)
			// fmt.Printf("part node: %v \n", current.Children[part])
		}
		current = current.Children[part]
		if i == len(parts)-1 {
			current.Value = val
			fullPath = fullPath + part
		} else {
			fullPath = fullPath + part + "/"
		}
		current.FullPath = fullPath
		// fmt.Printf("current node: %v \n", current.FullPath)
	}

}

func (n *Node) Ls(path string) []*Node {
	var parts []string
	// fmt.Printf("path inside ls: %v", path)
	parts = strings.Split(path, "/")

	current := n

	for i, part := range parts {
		if i == 0 {
			continue
		}
		child, exists := current.Children[part]
		if !exists {
			//fmt.Errorf("%v not valid child", part)
			wscutils.NewErrorResponse(" Insvalid child")
		}
		current = child
	}
	var nodes []*Node
	for _, v := range current.Children {
		nodes = append(nodes, v)
	}
	return nodes
}
