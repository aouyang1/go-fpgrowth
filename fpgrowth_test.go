package fpgrowth

import (
	"testing"
)

func TestNew(t *testing.T) {
	minSupport := 0.7
	fpg, err := New(minSupport)
	if err != nil {
		t.Error(err)
		return
	}
	if fpg.MinSupport != minSupport {
		t.Errorf("expected, %.3f, for min support but got %.3f", minSupport, fpg.MinSupport)
		return
	}
	if err := sameNode(fpg.tree, newNode(RootName)); err != nil {
		t.Error(err)
		return
	}
}

func testTransactions() []*Transaction {
	return []*Transaction{
		{0, []string{"f", "a", "c", "d", "g", "i", "m", "p"}},
		{1, []string{"a", "b", "c", "f", "l", "m", "o"}},
		{2, []string{"b", "f", "h", "j", "o"}},
		{3, []string{"b", "c", "k", "s", "p"}},
		{4, []string{"a", "f", "c", "e", "l", "p", "m", "n"}},
	}
}

func TestInsert(t *testing.T) {
	transactions := testTransactions()
	fpg, err := New(0.7)
	if err != nil {
		t.Error(err)
		return
	}
	for _, tr := range transactions {
		if err := fpg.insert(tr); err != nil {
			t.Error(err)
			return
		}
	}

	if len(transactions) != len(fpg.transactions) {
		t.Errorf("expected %d transactions but got %d", len(transactions), len(fpg.transactions))
		return
	}

	for i, trx := range transactions {
		if trx.ID != fpg.transactions[i].ID {
			t.Errorf("expected ID, %d for transaction at index %d, but got %d", fpg.transactions[i].ID, i, trx.ID)
			return
		}
		if len(trx.Items) != len(fpg.transactions[i].Items) {
			t.Errorf("expected %d items for transaction id %d, but got %d", len(trx.Items), trx.ID, len(fpg.transactions[i].Items))
			return
		}
		for j, item := range trx.Items {
			if item != fpg.transactions[i].Items[j] {
				t.Errorf("expected item, %s, but got %s", item, fpg.transactions[i].Items[j])
				return
			}
		}
	}

	var numItems int
	for _, trx := range transactions {
		numItems += len(trx.Items)
	}
	if fpg.frequentItems.n != numItems {
		t.Errorf("expected %d frequent item transactions, but got %d", numItems, fpg.frequentItems.n)
		return
	}

	expectedFI := map[string]*frequentItemCount{
		"f": {nil, 4},
		"c": {nil, 4},
		"p": {nil, 3},
		"m": {nil, 3},
		"b": {nil, 3},
		"a": {nil, 3},
	}

	for k, fic := range expectedFI {
		if val, exists := fpg.frequentItems.cnt[k]; !exists {
			t.Errorf("expected to find item, %s, in frequent item set", k)
			return
		} else {
			if val.count != fic.count {
				t.Errorf("expected count of %d, for item, %s, but got %d", fic.count, k, val.count)
				return
			}
		}
	}
}

func TestFit(t *testing.T) {
	transactions := testTransactions()

	fpg, err := New(0.09) // permits a count of 3 or more
	if err != nil {
		t.Error(err)
		return
	}
	if err := fpg.Fit(transactions); err != nil {
		t.Error(err)
		return
	}

	expectedTree := &node{
		item:  RootName,
		count: 0,
		children: map[string]*node{
			"f": {"f", 4, nil, nil,
				map[string]*node{
					"c": {"c", 3, nil, nil,
						map[string]*node{
							"p": {"p", 2, nil, nil,
								map[string]*node{
									"m": {"m", 2, nil, nil,
										map[string]*node{
											"a": {"a", 2, nil, nil, make(map[string]*node)},
										},
									},
								},
							},
							"m": {"m", 1, nil, nil,
								map[string]*node{
									"b": {"b", 1, nil, nil,
										map[string]*node{
											"a": {"a", 1, nil, nil, make(map[string]*node)},
										},
									},
								},
							},
						},
					},
					"b": {"b", 1, nil, nil, make(map[string]*node)},
				},
			},
			"c": {"c", 1, nil, nil,
				map[string]*node{
					"p": {"p", 1, nil, nil,
						map[string]*node{
							"b": {"b", 1, nil, nil, make(map[string]*node)},
						},
					},
				},
			},
		},
	}

	if err := compareTree(expectedTree, fpg.tree); err != nil {
		t.Error(err)
		return
	}

	expectedHeader := map[string][]int{
		"f": {4},
		"c": {3, 1},
		"p": {2, 1},
		"m": {2, 1},
		"b": {1, 1, 1},
		"a": {2, 1},
	}

	for item, counts := range expectedHeader {
		node := fpg.frequentItems.cnt[item].head
		for _, cnt := range counts {
			if node == nil {
				t.Errorf("expected initialization of header for item, %s", item)
				return
			}
			if node.count != cnt {
				t.Errorf("expected count, %d, for item, %s, but got %d", counts, item, node.count)
				return
			}
			node = node.next
		}
		if node != nil {
			t.Errorf("expected no more links for item, %s, but got, %v", item, node)
			return
		}
	}

	expectedPatternBases := []*patternBase{
		{"f", []itemCount{}},
		{"c", []itemCount{{"f", 3}}},
		{"p", []itemCount{{"c", 3}}},
		{"m", []itemCount{{"f", 3}, {"c", 3}}},
		{"b", []itemCount{}},
		{"a", []itemCount{{"f", 3}, {"c", 3}, {"m", 3}}},
	}
	if len(expectedPatternBases) != len(fpg.patternBases) {
		t.Errorf("expected %d pattern bases, but got %d", len(expectedPatternBases), len(fpg.patternBases))
		return
	}
	for i, pb := range fpg.patternBases {
		epb := expectedPatternBases[i]
		if pb.Item != epb.Item {
			t.Errorf("expected %s, but got %s for index, %d", epb.Item, pb.Item, i)
		}
		if len(pb.SubPatternBase) != len(epb.SubPatternBase) {
			t.Errorf("expected sub pattern base length of %d, but got %d for index, %d", len(epb.SubPatternBase), len(pb.SubPatternBase), i)
			continue
		}
		for j, item := range epb.SubPatternBase {
			if item.name != pb.SubPatternBase[j].name {
				t.Errorf("expected item, %s, but got %s, for index %d", item.name, pb.SubPatternBase[j].name, j)
			}
			if item.count != pb.SubPatternBase[j].count {
				t.Errorf("expected count, %d, but got %d, for index %d", item.count, pb.SubPatternBase[j].count, j)
			}

		}
	}
}

func TestFindPrefixPath(t *testing.T) {
	testData := []struct {
		n        *node
		expected []string
	}{
		{
			&node{item: "a", parent: &node{item: "b", parent: &node{item: "c"}}},
			[]string{"c", "b", "a"},
		},
		{
			&node{item: "a", parent: &node{item: "b", parent: &node{item: "c", parent: &node{item: RootName}}}},
			[]string{"c", "b", "a"},
		},
		{
			nil,
			[]string{},
		},
	}

	for _, td := range testData {
		res := findPrefixPath(td.n)
		if len(res) != len(td.expected) {
			t.Errorf("expected %d but got %d", len(td.expected), len(res))
			break
		}
		for i, item := range res {
			if item != td.expected[i] {
				t.Errorf("expected %s at index %d, but got %s", td.expected[i], i, item)
			}
		}
	}
}

func TestConditionalPatternBases(t *testing.T) {
	transactions := testTransactions()

	fpg, err := New(0.09) // permits a count of 3 or more
	if err != nil {
		t.Error(err)
		return
	}
	if err := fpg.Fit(transactions); err != nil {
		t.Error(err)
		return
	}

	item := fpg.frequentItems.itemCounts[len(fpg.frequentItems.itemCounts)-1].name
	expected := [][]itemCount{
		{{"f", 2}, {"c", 2}, {"p", 2}, {"m", 2}},
		{{"f", 1}, {"c", 1}, {"m", 1}, {"b", 1}},
	}
	cpb := fpg.conditionalPatternBases(item)
	if len(expected) != len(cpb) {
		t.Errorf("expected %d conditional base patterns, but got %d", len(expected), len(cpb))
		return
	}
	for i, pb := range cpb {
		if len(pb) != len(expected[i]) {
			t.Errorf("expected %d conditional pattern base items, but got %d", len(expected[i]), len(pb))
			break
		}
		for j, item := range pb {
			if item.name != expected[i][j].name {
				t.Errorf("expected %s, at index %d, but got %s", expected[i][j].name, j, item.name)
				break
			}
			if item.count != expected[i][j].count {
				t.Errorf("expected %d, at index %d, but got %d", expected[i][j].count, j, item.count)
				break
			}
		}
	}
}

func TestIntersectConditionalPatternBases(t *testing.T) {
	testData := []struct {
		cpb      [][]itemCount
		expected []itemCount
	}{
		{
			[][]itemCount{
				{{"a", 3}, {"b", 3}, {"c", 3}, {"d", 3}},
				{{"b", 2}, {"d", 2}},
			},
			[]itemCount{{"b", 5}, {"d", 5}},
		},
		{
			[][]itemCount{},
			[]itemCount{},
		},
		{
			[][]itemCount{
				{{"a", 2}, {"b", 2}, {"c", 2}, {"d", 2}},
			},
			[]itemCount{{"a", 2}, {"b", 2}, {"c", 2}, {"d", 2}},
		},
	}

	for _, td := range testData {
		res := intersectConditionalPatternBases(td.cpb)
		if len(res) != len(td.expected) {
			t.Errorf("expected %d, but got %d", len(td.expected), len(res))
			break
		}
		for i, item := range td.expected {
			if item.name != res[i].name {
				t.Errorf("expected %s, at index %d, but got %s", item.name, i, res[i].name)
			}
			if item.count != res[i].count {
				t.Errorf("expected %d, at index %d, but got %d", item.count, i, res[i].count)
			}

		}
	}
}
