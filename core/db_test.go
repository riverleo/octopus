package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDB(t *testing.T) {
	cachedDB = nil
	assert.Panics(t, func() { GetDB() }, "First use `func SetDB(...)` to set up the database.")

	assert.NotPanics(t, func() {
		SetTestDB()
		SetDBByEnv("test")
		GetDB()
		DropTestDB()
	})
}
