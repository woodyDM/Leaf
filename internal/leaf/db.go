package leaf

import (
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"path"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var Db *gorm.DB

func init() {
	home := GlobalConfig.Home
	dbPath := path.Join(home, "leaf.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to connect database %s", dbPath)))
	}
	log.Printf("Connect to %s success!", dbPath)
	Db = db
	//create table if need
	Db.AutoMigrate(&Application{})
	Db.AutoMigrate(&Task{})
	Db.AutoMigrate(&Env{})
	Db.AutoMigrate(&UsedEnv{})
	Db.AutoMigrate(&User{})
}

//todo delete
func RunSample() {

	// Migrate the schema


	// Create
	Db.Create(&Product{Code: "D42", Price: 100})

	// Read
	var product Product
	Db.First(&product, 1)                 // find product with integer primary key
	Db.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	Db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	Db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	Db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	//	db.Delete(&product, 1)
}


