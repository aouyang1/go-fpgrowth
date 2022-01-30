package fpgrowth

import "fmt"

type patternBase struct {
	Item           string
	SubPatternBase []itemCount
}

func (p *patternBase) String() string {
	return fmt.Sprintf("item: %s, subpattern: %v", p.Item, p.SubPatternBase)
}
