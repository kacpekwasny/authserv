package main

import (
	"fmt"
	"time"

	auth "github.com/kacpekwasny/authserv/src"
)

func main() {
	fmt.Println("server init...")
	var (
		root_url_to_api         = "/authserv/"
		mysqluser               = "authservuser1"
		password                = "Nyw)5(pjmL" // that is fine
		address                 = "10.8.0.1"   // server is not public, and just for this test
		port                    = 3306
		database_name           = "authserv"
		table_name              = "accounts1"
		ping_interval           = time.Second * 2
		ping_timout             = time.Second
		buffer_size             = 1000
		buffer_observe_interval = time.Second
	)
	s := auth.InitServer(root_url_to_api, mysqluser, password, address, port, database_name, table_name, ping_interval, ping_timout, buffer_size, buffer_observe_interval)
	s.Cnf.AddClient("admin", "admin", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	fmt.Println("server running")
	s.SetLogLevel(3)
	s.SetManagerLogLevel(1)
	s.ListenAndServe()
}
