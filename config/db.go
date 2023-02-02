package config

import (
	"log"
	"time"

	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/utils"
)

// Database holds the database configuration
type Database struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	Options         map[string][]string
	MaxIdleConn     int
	MaxOpenConn     int
	MaxConnLifetime time.Duration
}

var db Database

// DB returns the default database configuration
func DB() *Database {
	return &db
}

// LoadDB loads database configuration
func LoadDB() {
	log.Println("loading db")
	mu.Lock()
	defer mu.Unlock()

	username, password := utils.ParseDBCredentials()
	log.Println("username: ", username)
	log.Println("password: ", password)

	db = Database{
		Name:     "postgres",
		Username: username,
		Password: password,
		Host:     "postgres.db.svc",
		Port:     5432,
	}
}
