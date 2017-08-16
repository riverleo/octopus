package farmer

import (
	"github.com/finwhale/octopus/request"
)

func Exec(r *request.Request) interface{} {
	r.SetUp()

	if r.Operation == "query" {
		return Query(r.Node)
	}

	if r.Operation == "mutation" {
		return Mutation(r.Node)
	}

	return nil
}
