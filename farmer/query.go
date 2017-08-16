package farmer

import (
	"github.com/finwhale/octopus/request"
)

func Query(n *request.Node) interface{} {
	return n.Result()
}
