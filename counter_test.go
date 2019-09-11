package util

import (
	. "github.com/aandryashin/matchers"
	"testing"
)

func TestCounter(t *testing.T) {
	counter := NewCounter()
	AssertThat(t, counter.Count(), EqualTo{uint64(0)})
	AssertThat(t, counter.Get(), EqualTo{uint64(1)})
	AssertThat(t, counter.Count(), EqualTo{uint64(1)})
}
