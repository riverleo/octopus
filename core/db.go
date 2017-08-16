package core

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var cachedDB *gorm.DB

func GetDB() *gorm.DB {
	if cachedDB == nil {
		panic("First use `func SetDB(...)` to set up the database.")
	}

	return cachedDB
}

// 데이터베이스 인스턴스를 생성합니다.
func SetDB(adapter string, dbUrl string, schema string, charset string, maxOpenConns int, isPlural bool, isLogMode bool) *gorm.DB {
	var err error

	if adapter == "mysql" {
		cachedDB, err = gorm.Open(adapter, fmt.Sprintf("%v%v?charset=%v&parseTime=True&loc=Local", dbUrl, schema, charset))
		cachedDB.LogMode(isLogMode)
		cachedDB.SingularTable(!isPlural)
		Check(err)
	} else {
		// TODO will be support other adapters next time...
	}

	cachedDB.DB().SetMaxOpenConns(maxOpenConns)
	cachedDB.DB().SetMaxIdleConns(maxOpenConns)

	return cachedDB
}

func SetDBByEnv(env string) *gorm.DB {
	adapter, dbUrl, schema, charset, maxOpenConns, plural, logMode := GetSchemaInfo(env, true)
	db := SetDB(adapter, dbUrl, schema, charset, maxOpenConns, plural, logMode)

	return db
}

var lastEnv string

const testEnv = "test"

func SetTestDB() *gorm.DB {
	schema := GetSchema(false)
	if schema.Env != testEnv {
		lastEnv = schema.Env
	}

	adapter, dbUrl, schemaName, _, _, _, _ := GetSchemaInfo(testEnv, true)
	db, err := gorm.Open(adapter, dbUrl)
	Check(err)

	db.Exec("CREATE SCHEMA IF NOT EXISTS " + schemaName)
	db.Exec("USE " + schemaName)

	for _, table := range schema.Tables {
		db.Exec(table.CreateStatement())
	}

	db.Close()

	db = SetDBByEnv(testEnv)

	return db
}

func DropTestDB() {
	schema := GetSchema(false)
	adapter, dbUrl, schemaName, _, _, _, _ := GetSchemaInfo("test", true)
	db, err := gorm.Open(adapter, dbUrl+schemaName)
	defer db.Close()

	Check(err)

	for _, table := range schema.Tables {
		db.Exec(table.TruncateStatement())
	}
}
