package fpgrowth

import (
	"errors"
	"sort"
)

var (
	ErrNilTransaction    = errors.New("nil transaction")
	ErrInvalidMinSupport = errors.New("invalid minimum support. must be from 0 to 1.")
)

type FPGrowth struct {
	MinSupport float64

	frequentItems frequentItems
	transactions  []*Transaction // list of all transactions
	tree          *node
}

func New(minSupport float64) (*FPGrowth, error) {
	if minSupport > 1 || minSupport < 0 {
		return nil, ErrInvalidMinSupport
	}

	root := newNode("__root__")
	return &FPGrowth{
		MinSupport: minSupport,
		tree:       root,
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
				nextNode := newNode(i)
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

func (f *frequentItems) getSorted(minSupport float64) []string {
	if minSupport < 0 {
		minSupport = 0
	} else if minSupport > 1 {
		minSupport = 1
	}
	minCnt := int(minSupport * float64(f.n))

	f.itemCounts = f.itemCounts[:0]
	for itemName, fic := range f.cnt {
		if fic.count >= minCnt {
			f.itemCounts = append(f.itemCounts, itemCount{itemName, fic.count})
		}
	}
	sort.Slice(f.itemCounts, func(i, j int) bool {
		return f.itemCounts[i].count > f.itemCounts[j].count
	})

	items := make([]string, len(f.itemCounts))
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
		children: make(map[string]*node),
	}
}
