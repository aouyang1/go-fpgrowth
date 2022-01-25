package fpgrowth

import (
	"errors"
	"fmt"
)

type node struct {
	item     string
	count    int
	next     *node // points to another node with the same item name
	parents  map[string]*node
	children map[string]*node
}

func newNode(item string) *node {
	return &node{
		item:     item,
		parents:  make(map[string]*node),
		children: make(map[string]*node),
	}
}

func sameNode(a, b *node) error {
	if a == nil && b != nil {
		return errors.New("first argument is nil")
	}
	if a != nil && b == nil {
		return errors.New("second argument is nil")
	}
	if a == nil && b == nil {
		return nil
	}
	if a.item != b.item {
		return fmt.Errorf("expected item: %s, but got %s", a.item, b.item)
	}
	if a.count != b.count {
		return fmt.Errorf("expected count: %d, but got %d", a.count, b.count)
	}
	if len(a.parents) != len(b.parents) {
		return fmt.Errorf("expected parents map of size, %d, but got %d", len(a.parents), len(b.parents))
	}
	if len(a.children) != len(b.children) {
		return fmt.Errorf("expected children map of size, %d, but got %d", len(a.children), len(b.children))
	}
	return nil
}

// does a depth first search and compares node by node
func compareTree(a, b *node) error {
	if a == nil && b != nil {
		return errors.New("first argument is nil")
	}
	if a != nil && b == nil {
		return errors.New("second argument is nil")
	}
	if a == nil && b == nil {
		return nil
	}
	if a.item != b.item {
		return fmt.Errorf("expected item, %s, but got, %s", a.item, b.item)
	}
	if a.count != b.count {
		return fmt.Errorf("expected count, %d, for item, %s, but got %d", a.count, a.item, b.count)
	}
	if len(a.children) != len(b.children) {
		return fmt.Errorf("expected %d children, %v but got, %d, %v", len(a.children), a.children, len(b.children), b.children)
	}
	for item, node := range a.children {
		resnode, exists := b.children[item]
		if !exists {
			return fmt.Errorf("did not find item, %s, %v", item, b.children)
		}
		if err := compareTree(node, resnode); err != nil {
			return err
		}
	}
	return nil
}
