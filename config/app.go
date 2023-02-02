package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Application holds the application configuration
type Application struct {
	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// app is the default application configuration
var app Application

// App returns the default application configuration
func App() *Application {
	return &app
}

// LoadApp loads application configuration
func LoadApp() {
	log.Println("loading app")

	mu.Lock()
	defer mu.Unlock()

	app = Application{
		HTTPPort:     viper.GetInt("app.http_port"),
		ReadTimeout:  viper.GetDuration("app.read_timeout") * time.Second,
		WriteTimeout: viper.GetDuration("app.write_timeout") * time.Second,
		IdleTimeout:  viper.GetDuration("app.idle_timeout") * time.Second,
	}
}
