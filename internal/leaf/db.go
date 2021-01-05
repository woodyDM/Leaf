package leaf

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"path"
)

var Db *gorm.DB

type DbMode string

const (
	Sqlite DbMode = "sqlite"
	Mysql  DbMode = "mysql"
)

var DefaultDbConfig = DbConfig{
	Mode: Sqlite,
}

type DbConfig struct {
	Url  string `json:"url"`
	Mode DbMode `json:"mode"`
}

func InitDefault() error {
	return initDatabase(&DefaultDbConfig)
}
func InitWithConfig(path string) error {
	log.Printf("Reading configuration from file:[%s] ", path)
	bytes, e := ioutil.ReadFile(path)
	if e != nil {
		return e
	}
	var conf DbConfig
	e = json.Unmarshal(bytes, &conf)
	if e != nil {
		return e
	}
	if conf.Url == "" {
		return errors.New("Empty url .invalid config. ")
	}
	log.Printf("Configuration loaded from %s.", path)
	return initDatabase(&conf)
}

func initDatabase(c *DbConfig) error {
	home := GlobalConfig.Home
	var dbPath string
	var err error
	if c.Mode == Mysql {
		dbPath = c.Url
		Db, err = gorm.Open(mysql.Open(dbPath), &gorm.Config{})
	} else if c.Mode == Sqlite {
		dbPath = path.Join(home, "leaf.db")
		Db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	} else {
		return errors.New("Invalid db mode. ")
	}
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to connect database %s", dbPath)))
	}
	log.Printf(" Using [%s mode], Connect to %s success!", c.Mode, dbPath)
	//create table if need
	migrateTables()
	return nil
}

func migrateTables() {
	Db.AutoMigrate(&Application{})
	Db.AutoMigrate(&Task{})
	Db.AutoMigrate(&Env{})
	Db.AutoMigrate(&UsedEnv{})
	Db.AutoMigrate(&User{})
}
