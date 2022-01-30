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

	patternBases []*patternBase // stores pattern bases for each frequent item from most frequent to least
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
	f.patternBases = make([]*patternBase, len(f.frequentItems.itemCounts))
	for i := len(f.frequentItems.itemCounts) - 1; i >= 0; i-- {
		ic := f.frequentItems.itemCounts[i]
		cpb := f.conditionalPatternBases(ic.name)
		subpb := intersectConditionalPatternBases(cpb)
		f.patternBases[i] = &patternBase{
			Item:           ic.name,
			SubPatternBase: subpb,
		}
	}
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
				nextNode.parent = currNode

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

func (f *FPGrowth) conditionalPatternBases(item string) [][]itemCount {
	fi, exists := f.frequentItems.cnt[item]
	if !exists {
		return nil
	}

	var res [][]itemCount
	fip := fi.head

	for {
		if fip == nil {
			break
		}
		cpb := findPrefixPath(fip)
		if len(cpb) > 1 {
			items := make([]itemCount, 0, len(cpb)-1)
			for _, item := range cpb[:len(cpb)-1] {
				items = append(items, itemCount{item, fip.count})
			}
			res = append(res, items)
		}
		fip = fip.next
	}
	return res
}

func findPrefixPath(n *node) []string {
	if n == nil {
		return nil
	}
	if n.item == RootName {
		return nil
	}
	if n.parent == nil {
		return []string{n.item}
	}
	return append(findPrefixPath(n.parent), n.item)
}

func intersectConditionalPatternBases(cpb [][]itemCount) []itemCount {
	if len(cpb) == 0 {
		return nil
	}
	var fpSet map[string]int
	for _, pb := range cpb {
		pbSet := make(map[string]int)
		for _, item := range pb {
			pbSet[item.name] = item.count
		}
		if fpSet == nil {
			fpSet = pbSet
			continue
		}
		for item := range fpSet {
			if _, exists := pbSet[item]; !exists {
				delete(fpSet, item)
			} else {
				fpSet[item] += pbSet[item]
			}
		}
	}
	res := make([]itemCount, 0, len(fpSet))
	for _, item := range cpb[0] {
		if cnt, exists := fpSet[item.name]; exists {
			res = append(res, itemCount{item.name, cnt})
		}
	}
	return res
}
