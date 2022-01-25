package fpgrowth

import "sort"

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
	head  *node // points to first item in the FPTree and serves as the Header Table
	count int
}
