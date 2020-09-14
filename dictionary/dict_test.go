package dictionary

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeTickers(t *testing.T) {
	a := []Ticker{"A", "B", "C"}
	b := []Ticker{"A", "B", "D"}
	assert.Equal(t, []Ticker{"A", "B", "C", "D"}, MergeTickers(a, b))
}
