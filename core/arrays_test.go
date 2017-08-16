package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompact(t *testing.T) {
	arr := []string{"foo", "foo", "bar", "baz"}

	assert.Equal(t, Compact(arr), []interface{}{"foo", "bar", "baz"})
}

func TestCompactByPointer(t *testing.T) {
	arr := &[]string{"foo", "foo", "bar", "baz"}

	assert.Equal(t, Compact(arr), []interface{}{"foo", "bar", "baz"})
}
