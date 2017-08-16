package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	assert.Panics(t, func() { Check(fmt.Errorf("")) })
	assert.NotPanics(t, func() { Check(nil) })
}

func TestGetRootDir(t *testing.T) {
	assert.True(t, strings.HasSuffix(GetRootDir(), "/rice"))
}

func TestGetSchemaInfo(t *testing.T) {
	dbUrl, adapter, schema, charset, maxOpenConns, plural, logMode := GetSchemaInfo("test", true)
	assert.NotEmpty(t, dbUrl)
	assert.NotEmpty(t, adapter)
	assert.NotEmpty(t, schema)
	assert.NotEmpty(t, charset)
	assert.NotZero(t, maxOpenConns)
	assert.False(t, plural)
	assert.False(t, logMode)
}
