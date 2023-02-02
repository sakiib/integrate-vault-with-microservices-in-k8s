package config

import (
	"sync"
)

var mu sync.Mutex

// Init initiates of config load
func Init() {
	LoadApp()
	LoadDB()
}
