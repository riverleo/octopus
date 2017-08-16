package request

import (
	"fmt"
	"github.com/finwhale/octopus/core"
	"github.com/jinzhu/gorm"
	"math"
	"reflect"
	"strings"
)

// ------------------------------
// Set
// ------------------------------

func SetTotal(n *Node, _ *gorm.DB, data interface{}) {
	if !n.IsList || n.Find("_total") == nil {
		return
	}

	total := -1
	db := core.GetDB().Model(Get(n.Type))
	db.Count(&total)

	data.(map[string]interface{})["_total"] = total
}

func SetCount(n *Node, db *gorm.DB, data interface{}) {
	if !n.IsList || n.Find("_count") == nil {
		return
	}

	count := -1
	db.Offset(-1).Limit(-1).Count(&count)

	data.(map[string]interface{})["_count"] = count
}

func SetLimit(n *Node, _ *gorm.DB, data interface{}) {
	if !n.IsList || n.Find("_limit") == nil {
		return
	}

	if limit, exist := n.Args["_limit"]; exist {
		if limit == nil {
			data.(map[string]interface{})["_limit"] = core.GetConfig(false).Paging.Limit
		} else {
			data.(map[string]interface{})["_limit"] = limit
		}
		return
	}

	data.(map[string]interface{})["_limit"] = 0
}

func SetOffset(n *Node, _ *gorm.DB, data interface{}) {
	if !n.IsList || n.Find("_offset") == nil {
		return
	}

	if offset, exist := n.Args["_offset"]; exist {
		if offset == nil {
			data.(map[string]interface{})["_offset"] = core.GetConfig(false).Paging.Offset
		} else {
			data.(map[string]interface{})["_offset"] = offset
		}
		return
	}

	data.(map[string]interface{})["_offset"] = 0
}

// ------------------------------
// Query
// ------------------------------

func QueryLimitAndOffset(n *Node, db *gorm.DB) *gorm.DB {
	if !n.IsList {
		return db
	}

	limit := core.GetConfig(false).Paging.Limit
	maxLimit := core.GetConfig(false).Paging.MaxLimit
	offset := core.GetConfig(false).Paging.Offset

	if val, ok := n.Args["_limit"]; ok && core.IsKindOf(val, reflect.Float64) {
		limit = int(math.Min(val.(float64), float64(maxLimit)))
	}

	if val, ok := n.Args["_offset"]; ok && core.IsKindOf(val, reflect.Float64) {
		offset = int(math.Max(val.(float64), float64(offset)))
	}

	if limit <= 0 {
		db = db.Limit(maxLimit)
		db = db.Offset(offset)
	} else if limit > 0 {
		db = db.Limit(limit)
		db = db.Offset(offset)
	}

	return db
}

// 맵 프로퍼티를 판단하여 조인 쿼리를 생성합니다.
func QueryJoin(n *Node, db *gorm.DB) *gorm.DB {
	schema := core.GetSchema(false)

	for argName, argValue := range n.Args {
		if argValue == nil || reflect.TypeOf(argValue).Kind() != reflect.Map {
			continue
		}

		arg := argValue.(map[string]interface{})

		if len(arg) > 0 {
			joinTo := fmt.Sprintf("%v_id", argName)

			if v, exist := arg["_joinTo"].(string); exist {
				joinTo = v
			}

			originTable := schema.MustTable(n.Type)
			targetTable := schema.MustTable(arg["_tableName"].(string))
			originColumn := schema.MustColumn(originTable.Name, joinTo)
			targetPrimary, err := schema.GetPrimary(targetTable.Name)

			if joinFrom, exist := arg["_joinFrom"].(string); exist {
				targetPrimary = schema.MustColumn(targetTable.Name, joinFrom).Name
			}

			core.Check(err)

			db = db.Joins(
				fmt.Sprintf(
					"LEFT JOIN `%[2]v` `%[5]v` ON `%[1]v`.`%[3]v` = `%[5]v`.`%[4]v`",
					originTable.Name,
					targetTable.Name,
					originColumn.Name,
					targetPrimary,
					argName,
				),
			)

			for k, v := range arg {
				targetColumn := schema.GetColumn(targetTable.Name, k)

				if targetColumn == nil {
					continue
				}

				if core.IsKindOf(v, reflect.String) && (strings.HasPrefix(v.(string), "%") || strings.HasSuffix(v.(string), "%")) {
					db = db.Where(fmt.Sprintf("`%[1]v`.`%[2]v` LIKE ?", argName, targetColumn.Name), v)
				} else {
					db = db.Where(fmt.Sprintf("`%[1]v`.`%[2]v` = ?", argName, targetColumn.Name), v)
				}
			}
		}
	}

	return db
}
