package shortcuttrie

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/dfryer1193/mjolnir/utils/optional"
)

type ShortcutTrie struct {
	root *node
}

func (t *ShortcutTrie) String() string {
	return t.root.String()
}

type node struct {
	children map[string]*node
	shortcut optional.Option[string]
	varNode  *node
	argName  optional.Option[string]
}

func (n *node) String() string {
	return n.toStr(`""`, 0)
}

func (n *node) toStr(name string, level int) string {
	indent := strings.Repeat("  ", level)

	var argName string
	if n.argName.IsEmpty() {
		argName = `""`
	} else {
		argName = n.argName.Get()
	}

	var shortcut string
	if n.shortcut.IsEmpty() {
		shortcut = `""`
	} else {
		shortcut = n.shortcut.Get()
	}

	result := fmt.Sprintf("%sNode(name=%s, argName=%s, shortcut=%v)\n", indent, name, argName, shortcut)
	if n.varNode != nil {
		result += n.varNode.toStr("VAR", level+1)
	}
	for segment, child := range n.children {
		result += child.toStr(segment, level+1)
	}
	return result
}

func NewShortcutTrie() *ShortcutTrie {
	return &ShortcutTrie{
		root: &node{
			children: make(map[string]*node),
		},
	}
}

// From constructs a ShortcutTrie from an input reader. The input is expected to have one shortcut per line, with the format: /path/to/shortcut target_url
func From(input io.Reader) (*ShortcutTrie, error) {
	trie := NewShortcutTrie()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}
		path := parts[0]
		target := parts[1]
		trie.Insert(path, target)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}
	return trie, nil
}

// Insert adds a new shortcut path and its target to the trie. Paths are expecrted to include a leading slash, and wildcards (arguments) are wrapped with curly braces, e.g. /example/{arg}/test
// TODO: Ensure that path and target have already been unescaped with url.PathUnescape before calling.
func (t *ShortcutTrie) Insert(path string, target string) {
	current := t.root
	remainingPath := path
	nextSlashIndex := strings.Index(remainingPath[1:], "/")
	for len(remainingPath) > 0 {
		var segment string
		if nextSlashIndex == -1 {
			segment = remainingPath
			remainingPath = ""
			nextSlashIndex = -1
		} else {
			segment = remainingPath[:nextSlashIndex+1]
			remainingPath = remainingPath[nextSlashIndex+1:]
			nextSlashIndex = strings.Index(remainingPath[1:], "/")
		}

		next, exists := current.children[segment]
		if !exists {
			next = current.addChild(segment)
		}
		current = next
	}

	current.shortcut = optional.New(target)
}

// Get retrieves the target URL for a given shortcut path. It returns the target and a boolean indicating if the shortcut exists.
func (t *ShortcutTrie) Get(path string) (string, bool) {
	current := t.root
	remainingPath := path
	nextSlashIndex := strings.Index(remainingPath[1:], "/")
	varMap := make(map[string]string)
	for len(remainingPath) > 0 {
		var segment string
		if nextSlashIndex == -1 {
			segment = remainingPath
			remainingPath = ""
			nextSlashIndex = -1
		} else {
			segment = remainingPath[:nextSlashIndex+1]
			remainingPath = remainingPath[nextSlashIndex+1:]
			nextSlashIndex = strings.Index(remainingPath[1:], "/")
		}

		next, exists := current.children[segment]
		if !exists {
			if current.varNode != nil {
				next = current.varNode
				varMap[next.argName.Get()] = segment[1:]
			}
		}

		if next == nil {
			return "", false
		}

		current = next
	}

	if current.shortcut.IsEmpty() {
		return "", false
	}

	finalTarget, err := replaceArgs(current.shortcut.Get(), varMap)
	if err != nil {
		return "", false
	}
	return finalTarget, true
}

func (n *node) addChild(segment string) *node {
	var next *node
	isVariable := strings.HasPrefix(segment, "/{") && strings.HasSuffix(segment, "}")
	if isVariable {
		varNode := n.varNode
		argName := segment[2 : len(segment)-1]
		var shortcut optional.Option[string]
		if varNode != nil {
			shortcut = varNode.shortcut
		}
		next = &node{
			children: make(map[string]*node),
			argName:  optional.New(argName),
			shortcut: shortcut,
		}

		n.varNode = next
	} else {
		next = &node{
			children: make(map[string]*node),
		}

		n.children[segment] = next
	}

	return next
}

func replaceArgs(target string, varMap map[string]string) (string, error) {
	if len(varMap) == 0 {
		if strings.Contains(target, "{") {
			return "", fmt.Errorf("missing arguments for target: %s", target)
		}

		return target, nil
	}

	for arg, value := range varMap {
		placeholder := "{" + arg + "}"
		target = strings.ReplaceAll(target, placeholder, value)
	}

	if strings.Contains(target, "{") {
		return "", fmt.Errorf("missing arguments for target: %s", target)
	}

	return target, nil
}
