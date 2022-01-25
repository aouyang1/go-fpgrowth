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
	if err := sameNode(fpg.tree, newNode("__root__")); err != nil {
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
		if err := fpg.Insert(tr); err != nil {
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

func TestFrequentItemsGetSorted(t *testing.T) {
	transactions := testTransactions()

	fpg, err := New(0.09)
	if err != nil {
		t.Error(err)
		return
	}
	for _, tr := range transactions {
		if err := fpg.Insert(tr); err != nil {
			t.Error(err)
			return
		}
	}

	fi := fpg.frequentItems.getSorted(fpg.MinSupport)
	expected := []string{"f", "c", "p", "m", "b", "a"}
	if len(fi) != len(expected) {
		t.Errorf("expected %d items, but got %d", len(expected), len(fi))
		return
	}

	for i, item := range expected {
		if fi[i] != item {
			t.Errorf("expected item %s at index %d, but got %s", item, i, fi[i])
			return
		}
	}
}

func TestBuildTree(t *testing.T) {
	transactions := testTransactions()

	fpg, err := New(0.09) // permits a count of 3 or more
	if err != nil {
		t.Error(err)
		return
	}
	for _, tr := range transactions {
		if err := fpg.Insert(tr); err != nil {
			t.Error(err)
			return
		}
	}
	fpg.BuildTree()

	expectedTree := &node{
		item:  "__root__",
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
	}
}
