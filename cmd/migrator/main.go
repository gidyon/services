package main

import (
	"flag"
	"fmt"

	"github.com/gidyon/micro/v2/pkg/conn"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/internal/account"
)

var (
	dbHost     = flag.String("db-address", "localhost:3306", "MYSQL database address")
	dbUser     = flag.String("db-user", "root", "MYSQL database user")
	dbPassword = flag.String("db-password", "hakty11", "MYSQL database password")
	dbSchema   = flag.String("db-schema", "", "MYSQL database schema")
)

func main() {
	flag.Parse()

	db, err := conn.OpenGormConn(&conn.DBOptions{
		Dialect:  "mysql",
		Address:  *dbHost,
		User:     *dbUser,
		Password: *dbPassword,
		Schema:   *dbSchema,
	})
	errs.Panic(err)
	fmt.Println("STARTING AUTO MIGRATIONS")
	fmt.Println("...")
	errs.Panic(db.AutoMigrate(
		&account.Account{},
	))
	fmt.Println("DONE AUTOMIGRATIONS !!!!")
}
