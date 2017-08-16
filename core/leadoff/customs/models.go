package customs

import (
	"github.com/finwhale/octopus/request"
)

func init() {
	SetUp()
}

func SetUp() {
	request.Query = Query{}
	request.Mutation = Mutation{}
}

type (
	Query    struct{}
	Mutation struct{}
)
