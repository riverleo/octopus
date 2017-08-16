package core

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"path"
	"sort"
	"strings"
)

type (
	Schema struct {
		URL     string            `json:"url"`
		Adapter string            `json:"adapter"`
		Env     string            `json:"env"`
		Tables  map[string]*Table `json:"tables"`
	}

	Table struct {
		Name    string             `json:"name"`
		Columns map[string]*Column `json:"columns"`
	}

	Column struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Null    bool   `json:"null"`
		Key     string `json:"key"`
		Default string `json:"default"`
		Extra   string `json:"extra"`
	}

	ByLength []string
)

func (s ByLength) Len() int {
	return len(s)
}

func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

func (s *Schema) GetPrimary(tableName string) (name string, err error) {
	var names []string
	names, err = s.GetPrimaries(tableName)

	if len(names) > 0 {
		sort.Sort(ByLength(names))
		name = names[0]
	}

	return
}

func (s *Schema) GetPrimaries(tableName string) (names []string, err error) {
	table := s.GetTable(tableName)

	if table != nil {
		for _, column := range table.Columns {
			if column.Key == "PRI" {
				names = append(names, column.Name)
			}
		}
	}

	if len(names) == 0 {
		err = fmt.Errorf("%v is a nonexistent table or does not have a primary key.", tableName)
	}

	return
}

func (s *Schema) GetTable(name string) *Table {
	return s.Tables[CamelCase(name)]
}

func (s *Schema) MustTable(name string) *Table {
	table := s.GetTable(name)

	if table == nil {
		panic(fmt.Errorf("`%v` table does not exist.", name))
	}

	return table
}

func (s *Schema) GetColumn(tableName string, columnName string) *Column {
	table := s.GetTable(tableName)

	if table == nil {
		return nil
	}

	return table.Columns[CamelCase(columnName)]
}

func (s *Schema) MustColumn(tableName string, columnName string) *Column {
	column := s.GetColumn(tableName, columnName)

	if column == nil {
		panic(fmt.Errorf("`%v` column does not exist in `%v` table.", columnName, tableName))
	}

	return column
}

var cachedSchema *Schema

// 기존의 스키마를 가져옵니다. 없는 경우 schema.json 파일을 참조하여 신규 생성합니다.
func GetSchema(reload bool) *Schema {
	if !reload && cachedSchema != nil {
		return cachedSchema
	}

	path := path.Join(GetProjectDir(), DBFilename)
	if file, err := ioutil.ReadFile(path); err == nil {
		err = json.Unmarshal(file, &cachedSchema)
		Check(err)
	} else {
		cachedSchema = &Schema{}
	}

	return cachedSchema
}

// 타깃 데이터베이스의 스키마를 JSON 형태로 반환합니다.
func GetSchemaByDatabase(env string, adapter string, dbUrl string, schema string, charset string) *Schema {
	db, err := gorm.Open(adapter, fmt.Sprintf("%v%v?charset=%v&parseTime=True&loc=Local", dbUrl, schema, charset))
	Check(err)

	cachedSchema = &Schema{
		Env:     env,
		URL:     dbUrl,
		Adapter: adapter,
		Tables:  GetTables(db),
	}

	return cachedSchema
}

// 모든 테이블들을 불러옵니다.
func GetTables(db *gorm.DB) map[string]*Table {
	tableRows, err := db.Raw("show full tables where Table_Type = 'BASE TABLE'").Rows()
	defer tableRows.Close()
	Check(err)

	tables := map[string]*Table{}
	for tableRows.Next() {
		var tableName, tableType string
		tableRows.Scan(&tableName, tableType)
		table := &Table{Name: tableName}
		table.Columns = GetColumns(db, table)
		tables[CamelCase(tableName)] = table
	}

	return tables
}

// 해당 테이블의 모든 컬럼들을 불러옵니다.
func GetColumns(db *gorm.DB, table *Table) map[string]*Column {
	columnRows, err := db.Raw(fmt.Sprintf("DESC %v", table.Name)).Rows()
	defer columnRows.Close()
	Check(err)

	parseYes := func(yes string) bool {
		if strings.ToLower(yes) == "yes" {
			return true
		}

		return false
	}

	columns := map[string]*Column{}
	for columnRows.Next() {
		var cName, cType, cNull, cKey, cDefault, cExtra string
		columnRows.Scan(&cName, &cType, &cNull, &cKey, &cDefault, &cExtra)

		columns[CamelCase(cName)] = &Column{
			Name:    cName,
			Type:    cType,
			Null:    parseYes(cNull),
			Key:     cKey,
			Default: cDefault,
			Extra:   cExtra,
		}
	}

	return columns
}

func GetSchemaInfo(env string, reload bool) (adapter string, dbUrl string, schema string, charset string, maxOpenConns int, plural bool, logMode bool) {
	config, exist := GetConfig(reload).Database[env]

	if !exist {
		panic(fmt.Errorf("The schema of %v in the `config.yaml` file is not valid.", env))
	}

	username := config.Username
	password := config.Password
	database := config.Database
	maxOpenConns = config.MaxConnectionPool
	port := config.Port

	if username == "" {
		username = "root"
	}

	if password == "" {
		password = "root"
	}

	if database == "" {
		database = "127.0.0.1"
	}

	if maxOpenConns == 0 {
		maxOpenConns = 100
	}

	if port == "" {
		port = "3306"
	}

	dbUrl = fmt.Sprintf(
		"%v:%v@(%v:%v)/",
		username,
		password,
		config.Database,
		port,
	)
	adapter = config.Adapter
	charset = config.Charset
	schema = config.Schema
	plural = config.Plural
	logMode = config.LogMode

	return
}

func (t *Table) CreateStatement() string {
	createStatement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v (\n", t.Name)

	index := 0
	var primaries []string
	for _, column := range t.Columns {
		nullString := "NOT NULL"

		if column.Null {
			nullString = "NULL"
		}

		if column.Key == "PRI" {
			primaries = append(primaries, column.Name)

			if len(primaries) == 0 {
				createStatement += fmt.Sprintf("  `%v` %v NOT NULL AUTO_INCREMENT", column.Name, column.Type)
			} else {
				createStatement += fmt.Sprintf("  `%v` %v NOT NULL", column.Name, column.Type)
			}
		} else {
			createStatement += fmt.Sprintf("  `%v` %v %v", column.Name, column.Type, nullString)
		}

		if index != len(t.Columns)-1 {
			createStatement += ",\n"
		}

		index += 1
	}

	if len(primaries) > 0 {
		createStatement += fmt.Sprintf(",\n  PRIMARY KEY (%v)\n);", strings.Join(primaries, ","))
	} else {
		createStatement += ");"
	}

	return createStatement
}

func (t *Table) TruncateStatement() string {
	return fmt.Sprintf("TRUNCATE TABLE %v", t.Name)
}

func (t *Table) String() string {
	return t.Name
}

func (c *Column) String() string {
	return c.Name
}
