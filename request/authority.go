package request

import (
	"fmt"
	"github.com/finwhale/octopus/core"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"reflect"
	"regexp"
)

type (
	Authority struct {
		Default DefaultAuthority
		Models  map[string]AuthorityModel
	}

	DefaultAuthority struct {
		Read  Validator
		Write Validator
	}

	AuthorityModel struct {
		Read  Permission // 읽기 권한
		Write Permission // 쓰기 권한
	}

	Permission struct {
		Default Validator
		Fields  map[string][]Validator
	}

	Validator struct {
		Expression string   // 검증 방식 (hasId, hasRole...)
		Field      string   // 검증시 사용되는 데이터베이스 컬럼명, hasId 에서 사용됨
		Values     []string // 검증 시 사용되는 리터럴 값들, hasRole 에서 사용됨
	}
)

var cachedAuthority *Authority

const AuthorityFilename = "authority.yaml"

func GetAuthority(reload bool) *Authority {
	if reload || cachedAuthority == nil {
		file, err := ioutil.ReadFile(path.Join(core.GetProjectDir(), AuthorityFilename))

		if err != nil {
			return &Authority{}
		}

		var rawAuthority interface{}
		err = yaml.Unmarshal(file, &rawAuthority)
		core.Check(err)

		authority := parseAuthority(rawAuthority)
		cachedAuthority = &authority
	}

	if cachedAuthority == nil {
		panic(fmt.Errorf("Authority has not been correctly initialized."))
	}

	return cachedAuthority
}

// ------------------------------
// Authority
// ------------------------------

func (a *Authority) Analyze(n *Node) (validatorMap map[string][]Validator, persists []string) {
	isRead := core.Classify(n.Request.Operation) == "Query"
	isWrite := core.Classify(n.Request.Operation) == "Mutation"

	if isRead {
		return a.AnalyzeRead(n)
	}

	if isWrite {
		return a.AnalyzeWrite(n)
	}

	panic(fmt.Errorf("%v is an operation that can not be performed.", n.Request.Operation))
}

func (a *Authority) AnalyzeRead(n *Node) (validatorMap map[string][]Validator, fields []string) {
	var authorityModel *AuthorityModel

	// 해당 노드에 대해 설정된 검증 객체가 있는지 찾습니다.
	if model, exist := a.Models[n.Type]; exist {
		authorityModel = &model
	}

	// 노드에 대한 인증이 없는 경우 부모노드에 대해 찾습니다.
	if authorityModel == nil && n.Parent != nil {
		if model, exist := a.Models[n.Parent.Type]; exist {
			authorityModel = &model
		}
	}

	validatorMap = map[string][]Validator{}
	for _, child := range n.Fields {
		var validators []Validator

		if authorityModel == nil {
			validators = append(validators, a.Default.Read)
		} else if vds, exist := authorityModel.Read.Fields[child.Name]; exist {
			for _, vd := range vds {
				validators = append(validators, vd)
			}
		} else {
			// 아무런 검증 객체를 찾지 못한 경우 기본값으로 설정합니다.
			if _, exist := a.Models[child.Type]; !exist {
				validators = append(validators, authorityModel.Read.Default)
			}
		}

		validatorMap[child.Name] = validators
	}

	for _, validators := range validatorMap {
		for _, validator := range validators {
			if validator.Field == "" || core.Contains(fields, validator.Field) {
				continue
			}
			fields = append(fields, validator.Field)
		}
	}

	return
}

func (a *Authority) AnalyzeWrite(n *Node) (validatorMap map[string][]Validator, persists []string) {
	return
}

func parseAuthority(raw interface{}) Authority {
	a := Authority{}

	if raw == nil {
		return a
	}

	if core.IsKindOf(raw, reflect.Map) {
		rawMap := core.ParseMap(raw)

		if defaults, exist := rawMap["default"]; exist {
			if core.IsKindOf(defaults, reflect.String) {
				a.Default.Read = parseValidator(defaults.(string))
				a.Default.Write = parseValidator(defaults.(string))
			} else {
				defaultMap := core.ParseMap(defaults)

				if v, exist := defaultMap["read"]; exist {
					a.Default.Read = parseValidator(v.(string))
				}

				if v, exist := defaultMap["write"]; exist {
					a.Default.Write = parseValidator(v.(string))
				}
			}
		}

		if models, exist := rawMap["models"]; exist {
			m := core.ParseMap(models)
			a.Models = map[string]AuthorityModel{}

			for k, v := range m {
				a.Models[core.Classify(k)] = parseAuthorityModel(v, a.Default)
			}
		}
	} else {
		panic(fmt.Errorf("`%v` is incomprehensible structure.", raw))
	}

	return a
}

// ------------------------------
// AuthorityModel
// ------------------------------

func parseAuthorityModel(raw interface{}, defaults DefaultAuthority) AuthorityModel {
	a := AuthorityModel{Read: Permission{Default: defaults.Read}, Write: Permission{Default: defaults.Write}}

	if core.IsKindOf(raw, reflect.String) {
		a.Read = parsePermission(raw.(string), defaults.Read)
		a.Write = parsePermission(raw.(string), defaults.Write)
	} else if core.IsKindOf(raw, reflect.Map) {
		m := core.ParseMap(raw)

		if rawRead, exist := m["read"]; exist {
			a.Read = parsePermission(rawRead, defaults.Read)
		}

		if rawWrite, exist := m["write"]; exist {
			a.Write = parsePermission(rawWrite, defaults.Write)
		}
	}

	return a
}

// ------------------------------
// Permission
// ------------------------------

func parsePermission(raw interface{}, defaultValidator Validator) Permission {
	p := Permission{Default: defaultValidator}

	if core.IsKindOf(raw, reflect.String) {
		p.Default = parseValidator(raw.(string))
	} else if core.IsKindOf(raw, reflect.Map) {
		m := core.ParseMap(raw)

		if defaults, exist := m["default"]; exist {
			p.Default = parseValidator(defaults.(string))
		}

		if fields, exist := m["fields"]; exist {
			p.Fields = map[string][]Validator{}
			for k, v := range core.ParseMap(fields) {
				if core.IsKindOf(v, reflect.Slice) {
					vs := reflect.ValueOf(v)
					matchers := []Validator{}

					for i := 0; i < vs.Len(); i++ {
						m := vs.Index(i).Interface()
						matchers = append(matchers, parseValidator(m.(string)))
					}

					p.Fields[core.CamelCase(k)] = matchers
				} else if core.IsKindOf(v, reflect.String) {
					matchers := []Validator{parseValidator(v.(string))}
					p.Fields[core.CamelCase(k)] = matchers
				} else {
					panic(fmt.Errorf("`%v` can only be entered in string and array. (%v)", k, v))
				}
			}
		}
	}

	return p
}

// ------------------------------
// Validator
// ------------------------------

func (m Validator) IsAll() bool {
	return m.Expression == ""
}

func (m Validator) IsHasId() bool {
	return m.Expression == "hasId"
}

func (m Validator) IsHasRole() bool {
	return m.Expression == "hasRole"
}

func (m *Validator) Exec(n *Node, model interface{}) (statusCode int, errorMessage string) {
	field := reflect.Indirect(reflect.ValueOf(model)).FieldByName(core.Classify(n.Name))
	method := reflect.ValueOf(n.Request.GetUser()).MethodByName(core.Classify(m.Expression))

	if m.IsAll() {
		return 200, ""
	}

	if !method.IsValid() {
		panic(fmt.Errorf("%v - `%v` is invalid expression.", n.Name, m.Expression))
	}

	var args []reflect.Value
	if m.IsHasId() {
		args = append(args, reflect.ValueOf(field.Interface()))
	} else if m.IsHasRole() {
		for _, value := range m.Values {
			args = append(args, reflect.ValueOf(value))
		}
	}

	isValid := method.Call(args)[0].Bool()

	if isValid {
		return 200, ""
	}

	return 401, fmt.Sprintf("No permission to read `%v`.", n.Name)
}

func parseValidator(raw string) Validator {
	m := Validator{}
	fieldRegex := regexp.MustCompile("\\(\\.([a-zA-Z]+)\\)")
	valueRegex := regexp.MustCompile("\"([a-zA-Z]+)\"")
	expressRegex := regexp.MustCompile("^([a-zA-Z]+)")

	matchedFields := fieldRegex.FindStringSubmatch(raw)
	matchedValues := valueRegex.FindAllStringSubmatch(raw, -1)

	m.Expression = core.CamelCase(expressRegex.FindString(raw))

	if len(matchedFields) == 2 {
		m.Field = matchedFields[1]
	} else if m.IsHasId() {
		panic(fmt.Errorf("`%v` is incomprehensible expressions.", raw))
	}

	for _, v := range matchedValues {
		m.Values = append(m.Values, v[1])
	}

	if m.IsHasRole() && len(m.Values) == 0 {
		panic(fmt.Errorf("`%v` is incomprehensible expressions.", raw))
	}

	if !m.IsAll() && !m.IsHasId() && !m.IsHasRole() {
		panic(fmt.Errorf("`%v` is not support expressions.", raw))
	}

	return m
}
