package request

import (
	"encoding/json"
	"fmt"
	"github.com/finwhale/octopus/core"
	"github.com/jinzhu/gorm"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	GET              = "get"
	SCAN             = "scan"
	BULK             = "bulk"
	JOIN             = "join"
	OR               = "_or"
	AND              = "_and"
	WHERE            = "_where"
	ORDER            = "_order"
	COUNT            = "_count"
	LIMIT            = "_limit"
	OFFSET           = "_offset"
	DATA             = "_data"
	ERROR            = "_error"
	QUERY            = "Query"
	EQUAL            = "eq"
	NOT_EQUAL        = "ne"
	IN               = "in"
	NOT_IN           = "notIn"
	NIL              = "nil"
	LESS_THAN        = "lt"
	LESS_THAN_EQUAL  = "lte"
	GREAT_THAN       = "gt"
	GREAT_THAN_EQUAL = "gte"
	LIKE             = "like"
	INSENSITIVE_LIKE = "ilike"
	ASC              = "ASC"
	DESC             = "DESC"
	SUM              = "SUM"
	DEFAULT_USER     = "User"
	DATETIME         = "DateTime"
	FORMAT           = "format"
	KEY              = "key"
)

type (
	Join struct {
		Origin string
		Target string
	}

	Condition struct {
		Query string
		Args  []interface{}
	}

	Request struct {
		Name      string      `json:"name"`
		Operation string      `json:"operation"`
		user      CurrentUser `json:"-"`
		UserId    interface{} `json:"userId"`
		Node      *Node       `json:"node"`
		Header    http.Header `json:"-"`
	}

	Node struct {
		Name         string                 `json:"name"`
		Type         string                 `json:"type"`
		Args         map[string]interface{} `json:"args"`
		IsLeaf       bool                   `json:"isLeaf"`
		IsList       bool                   `json:"isList"`
		IsPlainList  bool                   `json:"isPlainList"`
		Fields       map[string]*Node       `json:"fields"`
		Parent       *Node                  `json:"-"`
		Request      *Request               `json:"-"` // 해당 노드가 포함된 요청
		Bulks        []string               `json:"-"` // List 형태의 출력에서 최적화된 벌크 컬럼의 이름들
		Customs      []string               `json:"-"` // 커스텀 메서드로 재정의된 컬럼의 이름들
		Scanneds     map[string][]string    `json:"-"` // 커스텀 또는 벌크 메서드에서 사용하려는 컬럼의 이름들
		Persists     []string               `json:"-"` // 데이터베이스에 존재하는 영속화된 컬럼의 이름들
		NoExists     []string               `json:"-"` // 커스텀, 모델, 벌크에서도 사용되지 않는 컬럼의 이름들
		Analyzed     bool                   `json:"-"` // 노드가 분석되었는지 여부
		Joins        []Join                 `json:"-"` // 조인이 필요한 테이블의 이름들
		Ors          [][]Condition          `json:"-"`
		Ands         [][][]Condition        `json:"-"`
		Wheres       []Condition            `json:"-"`
		Orders       []string               `json:"-"`
		ValidatorMap map[string][]Validator `json:"-"`
	}

	Result struct {
		DB   *gorm.DB
		Data interface{}
	}

	QueryHandler func(db *gorm.DB) *gorm.DB
)

// ------------------------------
// Request
// ------------------------------

func (r *Request) SetUp() {
	Connect(r.Node, r)
}

// 각 노드를 순회하면서 부모 노드 및 요청 객체와 연결합니다.
func Connect(n *Node, r *Request) {
	n.Request = r

	for _, field := range n.Fields {
		field.Parent = n
		field.Request = r
		Connect(field, r)
	}
}

func (r *Request) GetUser() CurrentUser {
	if r.user == nil {
		if r.UserId == "" {
			r.user = &AnonymousUser{}
		} else {
			userModelName := UserModelName

			// 설정된 사용자 모델명이 없다면 "User"를 기본값으로 사용한다.
			if userModelName == "" {
				userModelName = DEFAULT_USER
			}

			schema := core.GetSchema(false)
			table := schema.MustTable(userModelName)

			if table == nil {
				panic(fmt.Errorf("`%v` is not a valid model name", userModelName))
			}

			primary, err := schema.GetPrimary(table.Name)

			core.Check(err)

			user := New(userModelName, false)
			whereString := fmt.Sprintf("`%v`.`%v` = ?", table.Name, primary)
			scope := core.GetDB().Model(user).Where(whereString, r.UserId).First(user)

			if scope.RowsAffected > 0 {
				r.user = user.(CurrentUser)
			} else {
				r.user = &AnonymousUser{}
			}
		}
	}

	return r.user
}

// ------------------------------
// Node
// ------------------------------

// 노드가 요청하는 데이터를 GraphQL 규격에 맞게 변형하여 최종 결과물을 만듭니다.
func (n *Node) Result(handlers ...QueryHandler) *Result {
	n.Analyze(false)
	db, data := n.Fetch(n.IsList || n.IsPlainList, handlers...)

	if n.IsList {
		data = map[string]interface{}{
			DATA: data,
		}
	}

	SetTotal(n, db, data)
	SetCount(n, db, data)
	SetLimit(n, db, data)
	SetOffset(n, db, data)

	return &Result{
		DB:   db,
		Data: data,
	}
}

// 노드가 요청한 데이터를 변형하지 않고 불러온다.
func (n *Node) Fetch(isList bool, handlers ...QueryHandler) (*gorm.DB, interface{}) {
	n.Analyze(false)
	db, model := n.Query(isList, handlers...)
	fetchDB := db

	if _, exist := n.selectString(); exist {
		if isList {
			// 최상위 노드의 경우 동일한 아이템이 나오지 않도록 Group By 처리한다.
			if n.Parent == nil {
				schema := core.GetSchema(false)
				table := schema.MustTable(n.Type)
				primary, err := schema.GetPrimary(n.Type)

				if err == nil {
					fetchDB = fetchDB.Group(fmt.Sprintf("`%v`.`%v`", table.Name, primary))
				}
			}

			fetchDB = fetchDB.Find(model)
		} else {
			fetchDB = fetchDB.First(model)
		}
	}

	data := n.FulFill(fetchDB, model)
	data = n.Bulk(fetchDB, model, data)

	return db, data
}

func (n *Node) Query(isList bool, handlers ...QueryHandler) (db *gorm.DB, model interface{}) {
	model = Get(n.Type)

	if model == nil {
		panic(fmt.Errorf("`%v` is not support model type.", n.Type))
	}

	returnModel := New(n.Type, isList)
	method := reflect.ValueOf(model).MethodByName(QUERY)

	if method.IsValid() {
		args := []reflect.Value{}
		args = append(args, reflect.ValueOf(n))
		db = method.Call(args)[0].Interface().(*gorm.DB)
	} else {
		db = core.GetDB().Model(returnModel)
	}

	_select, _ := n.selectString()
	db = db.Select(_select)

	for _, join := range n.Joins {
		originModel := Get(join.Origin)
		method := reflect.ValueOf(originModel).MethodByName(core.EncapCase(JOIN, join.Target))

		if !method.IsValid() {
			panic(fmt.Errorf("`%v` model does not have a `%v` method.", join.Origin, core.EncapCase(JOIN, join.Target)))
		}

		values := []reflect.Value{reflect.ValueOf(db)}
		db = method.Call(values)[0].Interface().(*gorm.DB)
	}

	if len(n.Ors) > 0 {
		var query string
		var args []interface{}
		for i, conds := range n.Ors {
			var subQuery string

			for j, subConds := range conds {
				if j == len(conds)-1 {
					subQuery += subConds.Query
					args = append(args, subConds.Args...)
				} else {
					subQuery += subConds.Query + " AND "
					args = append(args, subConds.Args...)
				}
			}

			if i == len(n.Ors)-1 {
				query += "(" + subQuery + ")"
			} else {
				query += "(" + subQuery + ") OR "
			}
		}

		db = db.Where(query, args...)
	}

	if len(n.Ands) > 0 {
		var query string
		var args []interface{}

		for k, ors := range n.Ands {
			var orQuery string

			for i, conds := range ors {
				var subQuery string

				for j, subConds := range conds {
					if j == len(conds)-1 {
						subQuery += subConds.Query
						args = append(args, subConds.Args...)
					} else {
						subQuery += subConds.Query + " AND "
						args = append(args, subConds.Args...)
					}
				}

				if i == len(ors)-1 {
					orQuery += "(" + subQuery + ")"
				} else {
					orQuery += "(" + subQuery + ") OR "
				}
			}

			if k == len(n.Ands)-1 {
				query += "(" + orQuery + ")"
			} else {
				query += "(" + orQuery + ") AND "
			}
		}

		db = db.Where(query, args...)
	}

	for _, cond := range n.Wheres {
		db = db.Where(cond.Query, cond.Args...)
	}

	for _, order := range n.Orders {
		db = db.Order(order)
	}

	if isList {
		db = QueryLimitAndOffset(n, db)
	}

	for _, handler := range handlers {
		if handler != nil && reflect.TypeOf(handler).Kind() == reflect.Func {
			db = handler(db)
		}
	}

	return db, returnModel
}

// 해당 모델에서 정의된 커스텀 필드들을 호출
func (n *Node) FulFill(db *gorm.DB, model interface{}) interface{} {
	typeOf := reflect.TypeOf(model)

	// 배열 형태의 모델인 경우
	if typeOf.Kind() != reflect.Struct && typeOf.Elem().Kind() == reflect.Slice {
		data := []map[string]interface{}{}

		if db.RowsAffected > 0 {
			s := reflect.ValueOf(model).Elem()

			for i := 0; i < s.Len(); i++ {
				result := doFulFill(n, s.Index(i).Addr().Interface())
				data = append(data, result)
			}
		}

		return data
	}

	// 오브젝트 형태의 모델인 경우
	if db.RowsAffected > 0 {
		return doFulFill(n, model)
	}

	return nil
}

// 정의된 벌크 메서드를 실행하고 반영합니다. 리스트 형태의 요청에서만 동작합니다.
func (n *Node) Bulk(db *gorm.DB, models interface{}, data interface{}) interface{} {
	if db.RowsAffected == 0 || reflect.ValueOf(data).Kind() != reflect.Slice {
		return data
	}

	model := Get(n.Type)
	names := core.FlattenMap(n.Scanneds)

	extracted := core.GetByName(models, names...)

	for _, name := range n.Bulks {
		var args []reflect.Value
		args = append(args, reflect.ValueOf(n.Find(name)), reflect.ValueOf(extracted))
		method := reflect.ValueOf(model).MethodByName(core.EncapCase(BULK, name))
		called := method.Call(args)

		if called[1].IsValid() && called[1].Interface() != nil {
			panic(called[1])
		}

		bulked := called[0].Interface()
		n.merge(name, data, bulked)
	}

	return data
}

// GraphQL에서 클라이언트가 요청한 쿼리를 분석합니다.
func (n *Node) Analyze(force bool) (customs []string, persists []string, bulks []string, noexists []string, joins []Join) {
	model := Get(n.Type)
	schema := core.GetSchema(false)
	scanneds := map[string][]string{}

	// 올바른 노드가 아닌 경우 분석하지 않습니다.
	if model == nil {
		return
	}

	if n.Analyzed && !force {
		return n.Customs, n.Persists, n.Bulks, n.NoExists, n.Joins
	}

	n.ValidatorMap, persists = GetAuthority(false).Analyze(n)

	reflected := reflect.ValueOf(model)
	elemed := reflect.Indirect(reflected)

	var cJoins []Join

	if where, ok := n.Args[WHERE]; ok {
		wheres, joins := parseQuery(where, n, schema)
		n.Wheres = wheres
		cJoins = append(cJoins, joins...)
	}

	if order, ok := n.Args[ORDER]; ok {
		iterate(order, n.Type, func(isObject bool, parentName string, name string, source map[string]interface{}) {
			table := schema.GetTable(parentName)

			if table == nil {
				return
			}

			if isObject {
				childTable := schema.GetTable(name)

				if childTable == nil {
					return
				}

				cJoins = append(cJoins, Join{
					Origin: table.Name,
					Target: childTable.Name,
				})

				return
			}

			column := schema.GetColumn(table.Name, name)

			if column == nil {
				return
			}

			n.Orders = append(n.Orders, fmt.Sprintf("`%v`.`%v` %v", table.Name, column.Name, source["to"]))
		})
	}

	if orList, ok := n.Args[OR]; ok {
		orListValue := reflect.ValueOf(orList)

		for i := 0; i < orListValue.Len(); i++ {
			ors, joins := parseQuery(orListValue.Index(i).Interface(), n, schema)
			n.Ors = append(n.Ors, ors)
			cJoins = append(cJoins, joins...)
		}
	}

	if andList, ok := n.Args[AND]; ok {
		andListValue := reflect.ValueOf(andList)

		for i := 0; i < andListValue.Len(); i++ {
			orListValue := reflect.ValueOf(andListValue.Index(i).Interface())

			var orList [][]Condition
			for j := 0; j < orListValue.Len(); j++ {
				ors, joins := parseQuery(orListValue.Index(j).Interface(), n, schema)
				orList = append(orList, ors)
				cJoins = append(cJoins, joins...)
			}

			n.Ands = append(n.Ands, orList)
		}
	}

Loop:
	for _, cJoin := range cJoins {
		for _, join := range joins {
			if cJoin.Origin == join.Origin && cJoin.Target == join.Target {
				continue Loop
			}
		}

		joins = append(joins, cJoin)
	}

	// 사용자가 정의한 필드를 찾습니다.
	for _, field := range n.Fields {
		getMethod := reflected.MethodByName(core.EncapCase(GET, field.Name))
		scanMethod := reflected.MethodByName(core.EncapCase(SCAN, field.Name))
		bulkMethod := reflected.MethodByName(core.EncapCase(BULK, field.Name))

		if getMethod.IsValid() {
			customs = append(customs, field.Name)
		}

		if bulkMethod.IsValid() {
			bulks = append(bulks, field.Name)
		}

		if scanMethod.IsValid() {
			scannedValue := scanMethod.Call([]reflect.Value{})[0]

			if !scannedValue.CanInterface() {
				continue
			}

			scanned := scannedValue.Interface()
			returnKind := reflect.TypeOf(scanned).Elem().Kind()

			if scannedValue.Kind() == reflect.Slice && returnKind == reflect.String {
				scanneds[field.Name] = scanned.([]string)
			}
		}
	}

	// 하위 요소를 쿼리에서 가져올 때에는 반드시 `Primary`를 포함하도록 강제한다.
	primaries, err := schema.GetPrimaries(n.Type)
	if n.Parent != nil && err == nil {
		persists = append(persists, primaries...)
	}

	for _, field := range n.Fields {
		if elemed.FieldByName(core.Classify(field.Name)).IsValid() {
			if core.Contains(persists, field.Name) {
				continue
			}

			persists = append(persists, field.Name)
		} else if !core.Contains(customs, field.Name) {
			noexists = append(noexists, field.Name)
		}
	}

	n.Analyzed = true

	n.Joins = append(n.Joins, joins...)
	n.Bulks = append(n.Bulks, bulks...)
	n.Customs = append(n.Customs, customs...)
	n.Persists = append(n.Persists, persists...)
	n.NoExists = append(n.NoExists, noexists...)
	n.Scanneds = scanneds

	return
}

// 노드의 필드를 검색합니다.
func (n *Node) Find(candidate string) *Node {
	name := core.CamelCase(candidate)

	if strings.HasPrefix(candidate, "_") {
		name = candidate
	}

	if field, ok := n.Fields[name]; ok {
		return field
	}

	return nil
}

func (n *Node) Parse(value interface{}) interface{} {
	if n.Type == DATETIME {
		if format, exist := n.Args[FORMAT]; exist {
			t := value.(*time.Time)

			if t != nil {
				return t.Format(format.(string))
			}
		}
	}

	return value
}

func (n *Node) Empty() interface{} {
	isList := n.IsList || n.IsPlainList

	if isList {
		data := New(n.Type, isList)

		if n.IsList {
			return map[string]interface{}{
				DATA: data,
			}
		}

		return data
	}

	return nil
}

func (n *Node) merge(name string, data interface{}, bulked interface{}) interface{} {
	castedData := data.([]map[string]interface{})
	bulkedData := reflect.ValueOf(bulked)

	for i := 0; i < bulkedData.Len(); i++ {
		errors := castedData[i][ERROR].(map[string]interface{})[DATA].([]map[string]interface{})
		for _, err := range errors {
			if key, exist := err[KEY]; exist && key == name {
				return nil
			}
		}

		castedData[i][name] = bulkedData.Index(i).Interface()
	}

	return data
}

func (n *Node) selectString() (string, bool) {
	n.Analyze(false)

	schema := core.GetSchema(false)
	table := schema.MustTable(n.Type)
	conditions := []string{}

	for _, persist := range n.Persists {
		column := schema.MustColumn(n.Type, persist)
		name := fmt.Sprintf("`%v`.`%v`", table.Name, column.Name)

		if column == nil || core.Contains(n.Customs, name) {
			continue
		}

		conditions = append(conditions, name)
	}

	for _, scanned := range core.FlattenMap(n.Scanneds) {
		column := schema.GetColumn(n.Type, scanned)
		name := fmt.Sprintf("`%v`.`%v`", table.Name, column.Name)

		if column == nil || core.Contains(conditions, name) {
			continue
		}

		conditions = append(conditions, name)
	}

	// 스캔되지 않은 커스텀 필드가 있는 경우 모든 컬럼을 불러온다.
	existNotScannedCustom := false
	for _, custom := range n.Customs {
		if _, exist := n.Scanneds[custom]; !exist {
			existNotScannedCustom = true
			break
		}
	}

	if existNotScannedCustom {
		return "*", true
	}

	if len(conditions) == 0 {
		return "", false
	}

	return strings.Join(conditions, ", "), true
}

func (n *Node) Validate(candidate string, model interface{}) (errors []map[string]interface{}) {
	if validators, exist := n.ValidatorMap[candidate]; exist {
		for _, validator := range validators {
			statusCode, errorMessage := validator.Exec(n.Find(candidate), model)

			if errorMessage == "" {
				continue
			}

			err := map[string]interface{}{"key": candidate, "code": statusCode, "message": errorMessage}
			errors = append(errors, err)
		}
	}

	return errors
}

// ------------------------------
// Result
// ------------------------------

func (r *Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

// ------------------------------
// Utils
// ------------------------------

// 데이터 베이스의 노드 정보를 참고하여 모델의 데이터를 채워넣습니다.
func doFulFill(n *Node, model interface{}) map[string]interface{} {
	n.Analyze(false)
	value := reflect.ValueOf(model)
	elem := reflect.Indirect(value)
	field := elem.FieldByName("FulFilled")

	// 이미 FulFill이 실행된 모델에 대해 2번 반복하지 않습니다.
	if field.Len() > 0 {
		return field.Interface().(map[string]interface{})
	}

	errors := []map[string]interface{}{}
	fulfilled := map[string]interface{}{}

	for _, persist := range n.Persists {
		persistErrors := n.Validate(persist, model)
		errors = append(errors, persistErrors...)
		field := elem.FieldByName(core.Classify(persist))

		if len(errors) != 0 || !field.CanInterface() {
			continue
		}

		if fieldNode := n.Find(persist); fieldNode != nil {
			fulfilled[persist] = fieldNode.Parse(field.Interface())
		} else {
			fulfilled[persist] = field.Interface()
		}
	}

	for _, custom := range n.Customs {
		customErrors := n.Validate(custom, model)
		errors = append(errors, customErrors...)

		rootNode := n.Request.Node
		if len(errors) != 0 || (core.Contains(n.Bulks, custom) && (rootNode.IsList || rootNode.IsPlainList)) {
			continue
		}

		method := value.MethodByName(core.EncapCase(GET, custom))
		args := []reflect.Value{}
		args = append(args, reflect.ValueOf(n.Find(custom)))
		fulfilled[custom] = method.Call(args)[0].Interface()
	}

	errorMap := map[string]interface{}{}
	errorMap[DATA] = errors
	errorMap[COUNT] = len(errors)
	fulfilled[ERROR] = errorMap

	if field.CanSet() {
		field.Set(reflect.ValueOf(fulfilled))
	}

	return fulfilled
}

func iterate(
	c interface{},
	parentName string,
	creator func(isObject bool, parentName string, name string, source map[string]interface{}),
) {
	condition := reflect.ValueOf(c)
	objectKey := reflect.ValueOf("_object")

	if condition.Kind() != reflect.Map {
		panic(fmt.Errorf("Only the map type can be used."))
	}

	for _, nameValue := range condition.MapKeys() {
		name := nameValue.Interface().(string)
		sourceValue := reflect.ValueOf(condition.MapIndex(nameValue).Interface())

		if name == "_object" {
			continue
		}

		source := map[string]interface{}{}
		for _, sourceNameValue := range sourceValue.MapKeys() {
			sourceName := sourceNameValue.Interface().(string)
			source[sourceName] = sourceValue.MapIndex(sourceNameValue).Interface()
		}

		if sourceValue.MapIndex(objectKey).IsValid() {
			creator(true, parentName, name, source)
			iterate(sourceValue.Interface(), name, creator)

			continue
		}

		creator(false, parentName, name, source)
	}
}

func parseQuery(c interface{}, n *Node, schema *core.Schema) (conditions []Condition, joins []Join) {
	iterate(c, n.Type, func(isObject bool, parentName string, name string, source map[string]interface{}) {
		table := schema.GetTable(parentName)

		if table == nil {
			return
		}

		if isObject {
			childTable := schema.GetTable(name)

			if childTable == nil {
				return
			}

			joins = append(joins, Join{
				Origin: table.Name,
				Target: childTable.Name,
			})
			return
		}

		var queries []string
		var args []interface{}
		column := schema.GetColumn(table.Name, name)

		if column == nil {
			return
		}

		for opName, val := range source {
			if opName != NIL {
				args = append(args, val)
			}

			switch opName {
			case EQUAL:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` = ?", table.Name, column.Name))
				break
			case NOT_EQUAL:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` != ?", table.Name, column.Name))
				break
			case IN:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` IN (?)", table.Name, column.Name))
				break
			case NOT_IN:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` NOT IN (?)", table.Name, column.Name))
				break
			case NIL:
				if val.(bool) {
					queries = append(queries, fmt.Sprintf("`%v`.`%v` IS NULL", table.Name, column.Name))
				} else {
					queries = append(queries, fmt.Sprintf("`%v`.`%v` IS NOT NULL", table.Name, column.Name))
				}
			case LESS_THAN:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` < ?", table.Name, column.Name))
				break
			case LESS_THAN_EQUAL:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` <= ?", table.Name, column.Name))
				break
			case GREAT_THAN:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` > ?", table.Name, column.Name))
				break
			case GREAT_THAN_EQUAL:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` >= ?", table.Name, column.Name))
				break
			case LIKE:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` LIKE ?", table.Name, column.Name))
				break
			case INSENSITIVE_LIKE:
				queries = append(queries, fmt.Sprintf("`%v`.`%v` ILIKE ?", table.Name, column.Name))
				break
			default:
				continue
			}
		}

		conditions = append(conditions, Condition{
			Query: strings.Join(queries, " AND "),
			Args:  args,
		})
	})

	return
}
