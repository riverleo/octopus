package request

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetUp(t *testing.T) {
	r := &Request{
		Name:      "anonymous",
		Operation: "query",
		Node: &Node{
			Name: "user",
			Type: "User",
			Fields: map[string]*Node{
				"id":       &Node{Name: "id", Type: "Int"},
				"password": &Node{Name: "password", Type: "String"},
			},
		},
	}

	assert.Nil(t, r.Node.Request)
	assert.Nil(t, r.Node.Fields["id"].Parent)
	assert.Nil(t, r.Node.Fields["id"].Request)
	assert.Nil(t, r.Node.Fields["password"].Parent)
	assert.Nil(t, r.Node.Fields["password"].Request)

	r.SetUp()

	assert.Equal(t, r.Node.Request, r)
	assert.Equal(t, r.Node.Fields["id"].Parent, r.Node)
	assert.Equal(t, r.Node.Fields["id"].Request, r)
	assert.Equal(t, r.Node.Fields["password"].Parent, r.Node)
	assert.Equal(t, r.Node.Fields["password"].Request, r)
}

func TestNode_Analyze_Customs(t *testing.T) {
}

func TestNode_Analyze_Persists(t *testing.T) {
}

func TestNode_Analyze_Bulk(t *testing.T) {
}

func TestNode_Analyze_Noexists(t *testing.T) {
}

func GetTestRequest() *Request {
	GetFunc = func(_ string) interface{} {
		return struct {
			Name string
		}{
			"User",
		}
	}

	r := &Request{
		Name:      "anonymous",
		Operation: "query",
		Node: &Node{
			Name: "user",
			Type: "User",
			Args: map[string]interface{}{
				"_or": []map[string]interface{}{
					map[string]interface{}{
						"_object": true,
						"role": map[string]interface{}{
							"_object": true,
							"createdAt": map[string]string{
								EQUAL: "2017-6-17",
							},
							"roleType": map[string]interface{}{
								"_object": true,
								"name": map[string]string{
									NOT_EQUAL: "ADMIN",
								},
							},
						},
					},
				},
				"_where": map[string]interface{}{
					"_object": true,
					"name": map[string]string{
						EQUAL: "Leo",
					},
					"role": map[string]interface{}{
						"_object": true,
						"userId": map[string]interface{}{
							NOT_EQUAL:  3,
							GREAT_THAN: 39,
						},
						"createdAt": map[string]string{
							EQUAL: "2014-04-24",
						},
						"roleType": map[string]interface{}{
							"_object": true,
							"name": map[string]string{
								EQUAL: "ADMIN",
							},
						},
					},
				},
				"_order": map[string]interface{}{
					"name": map[string]string{
						"to": ASC,
					},
					"article": map[string]interface{}{
						"text": map[string]string{
							"to":   DESC,
							"func": SUM,
						},
						"comment": map[string]interface{}{
							"_object": true,
						},
						"_object": true,
					},
					"_object": true,
				},
			},
		},
	}

	r.SetUp()
	r.Node.Analyze(true)

	return r
}

func TestNode_Analyze(t *testing.T) {
	r := GetTestRequest()

	r.Node.Analyze(false)

	assert.Contains(t, r.Node.Wheres, Condition{Query: "`role_type`.`name` = ?", Args: []interface{}{"ADMIN"}})
	assert.Contains(t, r.Node.Wheres, Condition{Query: "`role`.`created_at` = ?", Args: []interface{}{"2014-04-24"}})

	assert.Contains(t, r.Node.Joins, Join{Origin: "user", Target: "role"})
	assert.Contains(t, r.Node.Joins, Join{Origin: "user", Target: "article"})
	assert.Contains(t, r.Node.Joins, Join{Origin: "role", Target: "role_type"})

	assert.Contains(t, r.Node.Orders, "`user`.`name` ASC")
	assert.Contains(t, r.Node.Orders, "`article`.`text` DESC")
	assert.Contains(t, r.Node.Ors[0], Condition{Query: "`role`.`created_at` = ?", Args: []interface{}{"2017-6-17"}})
	assert.Contains(t, r.Node.Ors[0], Condition{Query: "`role_type`.`name` = ?", Args: []interface{}{"ADMIN"}})
}
