package fpgrowth

import (
	"errors"
)

var (
	ErrNilTransaction    = errors.New("nil transaction")
	ErrInvalidMinSupport = errors.New("invalid minimum support. must be from 0 to 1.")
	RootName             = "__ROOT__"
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
		tree:          newNode(RootName),
		frequentItems: newFrequentItems(),
	}, nil
}

func (f *FPGrowth) Fit(t []*Transaction) error {
	for _, tx := range t {
		if err := f.insert(tx); err != nil {
			return err
		}
	}
	f.buildTree()
	return nil
}

func (f *FPGrowth) insert(t *Transaction) error {
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

func (f *FPGrowth) buildTree() {
	// find frequent items in sorted order
	fi := f.frequentItems.getSorted(f.MinSupport)

	// header table workspace
	headerTbl := make(map[string]*node)

	// second pass of transactions building th FP Tree
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

				// update header table
				if f.frequentItems.cnt[i].head == nil {
					f.frequentItems.cnt[i].head = nextNode
				}

				// link newly created nodes to header
				node, exists := headerTbl[i]
				if exists {
					node.next = nextNode
				}
				headerTbl[i] = nextNode
			}
			nextNode.count += 1
			currNode = nextNode
		}
	}
}
