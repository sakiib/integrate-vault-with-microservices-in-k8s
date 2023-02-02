package conn

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sakiib/integrate-vault-with-micriservices-in-k8s/config"
)

// DB holds the database instance
var db *gorm.DB

// Ping tests if db connection is alive
func Ping() error {
	return db.Exec("select id from users limit 1;").Error
}

// Connect sets the db client of database using configuration cfg
func Connect(cfg *config.Database) error {
	host := cfg.Host
	if cfg.Port != 0 {
		host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	uri := url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   cfg.Name,
		User:   url.UserPassword(cfg.Username, cfg.Password),
	}
	if cfg.Options != nil {
		val := url.Values(cfg.Options)
		uri.RawQuery = val.Encode()
	}
	// open a database connection using gorm ORM
	d, err := gorm.Open(postgres.Open(uri.String()), &gorm.Config{})
	if err != nil {
		return err
	}
	db = d

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if cfg.MaxIdleConn != 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	}
	if cfg.MaxOpenConn != 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	}
	if cfg.MaxConnLifetime.Seconds() != 0 {
		sqlDB.SetConnMaxLifetime(cfg.MaxConnLifetime)
	}

	return nil
}

// DefaultDB returns default db
func DefaultDB() *gorm.DB {
	return db.Debug()
}

// CloseDB closes the db connection
func CloseDB() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// ConnectDB sets the db client of database using default configuration file
func ConnectDB() error {
	cfg := config.DB()
	connectionRenew() //start a connection re-newer
	return Connect(cfg)
}

func connectionRenew() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if err := Ping(); err != nil {
					log.Printf("error: %v [re-connecting database]", err.Error())
					Connect(config.DB())
					_ = t
				} else {
					log.Println("db is still connected")
				}
			}
		}
	}()
}
