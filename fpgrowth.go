package fpgrowth

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrNilTransaction    = errors.New("nil transaction")
	ErrInvalidMinSupport = errors.New("invalid minimum support. must be from 0 to 1.")
)

type FPGrowth struct {
	MinSupport float64

	frequentItems *frequentItems
	transactions  []*Transaction // list of all transactions
	tree          *node
}

func New(minSupport float64) (*FPGrowth, error) {
	if minSupport > 1 || minSupport < 0 {
		return nil, ErrInvalidMinSupport
	}

	return &FPGrowth{
		MinSupport:    minSupport,
		tree:          newNode("__root__"),
		frequentItems: newFrequentItems(),
	}, nil
}

func (f *FPGrowth) Insert(t *Transaction) error {
	if t == nil {
		return ErrNilTransaction
	}

	f.transactions = append(f.transactions, t)
	for _, item := range t.Items {
		// only add items that are non empty strings
		if item != "" {
			f.frequentItems.add(item)
		}
	}
	return nil
}

func (f *FPGrowth) BuildTree() {
	fi := f.frequentItems.getSorted(f.MinSupport)
	for _, t := range f.transactions {
		currNode := f.tree
		for _, i := range fi {
			if !t.Exists(i) {
				continue
			}
			nextNode, ok := currNode.children[i]
			if !ok {
				nextNode = newNode(i)
				currNode.children[i] = nextNode
				nextNode.parents[currNode.item] = currNode
			}
			nextNode.count += 1
			currNode = nextNode
		}
	}
}

type frequentItems struct {
	n          int                           // number of items stored
	cnt        map[string]*frequentItemCount // tracks most frequent items of all transactions
	itemCounts itemCounts                    // item names sorted by most frequent
}

func newFrequentItems() *frequentItems {
	return &frequentItems{
		cnt: make(map[string]*frequentItemCount),
	}
}

func (f *frequentItems) reset() {
	f.n = 0
	f.cnt = make(map[string]*frequentItemCount)
	f.itemCounts = f.itemCounts[:0]
}

func (f *frequentItems) add(item string) {
	f.n += 1
	if fic, ok := f.cnt[item]; ok {
		fic.count += 1
	} else {
		f.cnt[item] = &frequentItemCount{nil, 1}
	}
}

func (f *frequentItems) get(item string) int {
	fic, ok := f.cnt[item]
	if ok {
		return fic.count
	}
	return 0
}

// return a sorted list of items based on high frequency and limited to min support
func (f *frequentItems) getSorted(minSupport float64) []string {
	if minSupport < 0 {
		minSupport = 0
	} else if minSupport > 1 {
		minSupport = 1
	}
	minCnt := minSupport * float64(f.n)

	f.itemCounts = f.itemCounts[:0]
	for itemName, fic := range f.cnt {
		if float64(fic.count) >= minCnt {
			f.itemCounts = append(f.itemCounts, itemCount{itemName, fic.count})
		}
	}
	sort.Slice(f.itemCounts, func(i, j int) bool {
		if f.itemCounts[i].count > f.itemCounts[j].count {
			return true
		}
		if f.itemCounts[i].count < f.itemCounts[j].count {
			return false
		}
		return f.itemCounts[i].name > f.itemCounts[j].name
	})

	items := make([]string, 0, len(f.itemCounts))
	for _, ic := range f.itemCounts {
		items = append(items, ic.name)
	}

	return items
}

type itemCounts []itemCount

type itemCount struct {
	name  string
	count int
}

type frequentItemCount struct {
	head  *node // points to first item
	count int
}

type Transaction struct {
	ID    int
	Items []string // slice of item names
}

func (t *Transaction) Exists(item string) bool {
	for _, i := range t.Items {
		if i == item {
			return true
		}
	}
	return false
}

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
