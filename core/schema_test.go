package core

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type SchemaSuite struct {
	suite.Suite
	schema string
}

func (s *SchemaSuite) TestGetPrimary() {
	schema := Schema{
		Tables: map[string]*Table{
			"foo": &Table{
				Name: "foo",
				Columns: map[string]*Column{
					"id": &Column{
						Name: "id",
						Key:  "PRI",
					},
				},
			},
		},
	}
	name, _ := schema.GetPrimary("foo")
	s.Equal(name, "id")
}

func (s *SchemaSuite) TestGetSchema() {
	var schema *Schema

	s.NotPanics(func() { schema = GetSchema(true) })
	s.Equal(cachedSchema, GetSchema(true))
}

func (s *SchemaSuite) TestCreateStatement() {
	schema := Schema{
		Tables: map[string]*Table{
			"foo": &Table{
				Name: "foo",
				Columns: map[string]*Column{
					"id": &Column{
						Name: "id",
						Key:  "PRI",
					},
				},
			},
		},
	}

	for _, table := range schema.Tables {
		s.NotEmpty(table.CreateStatement())
	}
}

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaSuite))
}
