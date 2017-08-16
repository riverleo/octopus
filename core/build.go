package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DBFilename        = "db.json"
	ModelFilename     = "models.go"
	ConfigFilename    = "config.yaml"
	AuthorityFilename = "authority.yaml"
)

func Build(makeFile bool, env string, adapter string, dbUrl string, schemaName string, charset string) *Schema {
	schema := GetSchemaByDatabase(env, adapter, dbUrl, schemaName, charset)

	if makeFile {
		// schema.json
		sBytes, _ := json.Marshal(schema)
		SaveToFile(DBFilename, sBytes, true)

		// models.go
		SaveToFile("./models/"+ModelFilename, []byte(printSchemaFile(schema)), true)
	}

	return schema
}

func SaveToFile(relativePath string, body []byte, overwrite bool) {
	pwd, err := os.Getwd()
	Check(err)

	absolutePath := path.Join(pwd, relativePath)
	if _, err = os.Stat(absolutePath); !overwrite && err == nil {
		return
	}

	dir := filepath.Dir(absolutePath)
	os.MkdirAll(dir, 0777)
	f, err := os.Create(absolutePath)
	defer f.Close()

	Check(err)
	f.Write(body)
	f.Sync()
}

func printSchemaFile(schema *Schema) string {
	template := "package models\n\nimport (\n  \"database/sql\"\n  \"encoding/json\"\n  \"fmt\"\n  \"github.com/wanteddev/rice/core\"\n  \"github.com/wanteddev/rice/request\"\n  \"time\"\n)\n\ntype NewFuncType func(isList bool) interface{}\n\nvar (\n  Env string\n  cachedModels map[string]interface{}\n  cachedNewFuncs map[string]NewFuncType\n)\n\nfunc SetUp(env string) {\n  Env = env\n  cachedModels = GetAll()\n  request.NewFunc = New\n  request.GetFunc = Get\n  request.GetAllFunc = GetAll\n}\n\nfunc Get(candidate string) interface{} {\n  return cachedModels[core.Classify(candidate)]\n}\n\nfunc GetAll() map[string]interface{} {\n  if len(cachedModels) == 0 {\n    cachedModels = map[string]interface{}{\n%[1]v    }\n  }\n  return cachedModels\n}\n\nfunc New(candidate string, isList bool) interface{} {\n  if len(cachedNewFuncs) == 0 {\n    cachedNewFuncs = map[string]NewFuncType {\n%[5]v    }  \n  }\n\n  if f, ok := cachedNewFuncs[candidate]; ok {\n    return f(isList)\n  }\n\n  return nil\n}\n\n%[3]v%[2]v%[4]v"

	index := 0
	var mapTemplate, newTemplate, newFuncTemplate, modelTemplate string
	for _, table := range schema.Tables {
		tableName := Classify(table.Name)
		tableString := fmt.Sprintf("type %v struct {\n  FulFilled map[string]interface{} `json\"-\",structs:\"-\",gorm:\"-\"`\n", tableName)

		for _, column := range table.Columns {
			columnType := "string"

			if column.Null {
				columnType = "NullString"
			}
			columnName := Classify(column.Name)
			gormTag := fmt.Sprintf("type:%v;column:%v", column.Type, column.Name)
			jsonTag := CamelCase(column.Name)

			if strings.HasPrefix(column.Type, "int") {
				columnType = "int64"

				if column.Null {
					columnType = "NullInt64"
				}
			} else if strings.HasPrefix(column.Type, "float") {
				columnType = "float64"

				if column.Null {
					columnType = "NullFloat64"
				}
			} else if strings.HasPrefix(column.Type, "tinyint(1)") {
				columnType = "bool"

				if column.Null {
					columnType = "NullBool"
				}
			} else if strings.HasPrefix(column.Type, "date") || strings.HasPrefix(column.Type, "time") {
				columnType = "*time.Time"
			}

			if column.Key == "PRI" {
				gormTag += ";primary_key"
			} else if column.Key == "UNI" {
				gormTag += ";unique_index"
			}

			if !column.Null {
				gormTag += ";not null"
			}

			tableString += fmt.Sprintf(
				"  %[1]v %[2]v `json:\"%[3]v\",structs:\"%[3]v\",gorm:\"%[4]v\"`\n",
				columnName, columnType, jsonTag, gormTag,
			)
		}

		mapTemplate += fmt.Sprintf("      \"%[1]v\": &%[1]v{},\n", tableName)
		newTemplate += fmt.Sprintf("      \"%[1]v\": New%[1]v,\n", tableName)
		newFuncTemplate += fmt.Sprintf("func New%[1]v(isList bool) interface{} {\n  if isList {\n    return &[]%[1]v{}\n  }\n  return &%[1]v{}\n}\n\n", tableName)
		modelTemplate += tableString + "}\n\n"
		index += 1
	}

	nullValueTemplate := `
type NullInt64 struct {
	sql.NullInt64
}

func (r NullInt64) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (r NullInt64) ToInt64() int64 {
	return r.Int64
}

func (r NullInt64) String() string {
	return fmt.Sprintf("%v", r.Int64)
}

type NullFloat64 struct {
	sql.NullFloat64
}

func (r NullFloat64) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (r NullFloat64) ToFloat64() float64 {
	return r.Float64
}

func (r NullFloat64) String() string {
	return fmt.Sprintf("%v", r.Float64)
}

type NullBool struct {
	sql.NullBool
}

func (r NullBool) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Bool)
	} else {
		return json.Marshal(nil)
	}
}

func (r NullBool) ToBool() bool {
	return r.Bool
}

func (r NullBool) String() string {
	return fmt.Sprintf("%v", r.Bool)
}

type NullString struct {
	sql.NullString
}

func (r NullString) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.NullString.String)
	} else {
		return json.Marshal(nil)
	}
}

func (r NullString) String() string {
	return r.NullString.String
}`

	return fmt.Sprintf(template, mapTemplate, modelTemplate, newFuncTemplate, nullValueTemplate, newTemplate)
}
