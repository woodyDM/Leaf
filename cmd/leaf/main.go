package main

import (
	"flag"
	"leaf/internal/leaf"
	"log"
)

var mode = flag.String("m", "server", "server: start server\nuser : create or update user")
var port = flag.Int("s", 8080, "server port (in server mode only)")
var user = flag.String("u", "admin", "user name (in user user only)")
var pass = flag.String("p", "123456", "user pass (in user user only)")
var configFile = flag.String("c", "", "path of config file ,if empty, use default db : sqlite)")

func main() {
	flag.Parse()
	var err error
	if *configFile == "" {
		err = leaf.InitDefault()
	} else {
		err = leaf.InitWithConfig(*configFile)
	}
	if err != nil {
		log.Fatalf("Unable to connect to db: %v", err)
	}
	if *mode == "server" {
		leaf.StartServer(*port)
	} else {
		err = leaf.FreshUser(*user, *pass)
		if err != nil {
			log.Println("Unable to fresh user! Reason is", err.Error())
		} else {
			log.Printf("Success fresh user %s\n", *user)
		}
	}
}
