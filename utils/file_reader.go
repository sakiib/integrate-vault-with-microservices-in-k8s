package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	username string
	password string
)

func ParseDBCredentials() (string, string) {
	file, err := os.Open("/vault/secrets/db-creds")
	if err != nil {
		log.Println(err)
		return "", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K
	for scanner.Scan() {
		temp := scanner.Text()
		fmt.Println("file temp: ", temp)
		if strings.HasPrefix(temp, "username=") {
			username = strings.TrimPrefix(temp, "username=")
		}
		if strings.HasPrefix(temp, "password=") {
			password = strings.TrimPrefix(temp, "password=")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return username, password
}
