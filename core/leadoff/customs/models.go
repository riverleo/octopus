package customs

import (
	"github.com/wanteddev/rice/request"
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
