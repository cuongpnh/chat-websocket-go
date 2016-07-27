package env

import (
	"os"
	"strconv"
)

func Get(name string) string {
	return os.Getenv(name)
}

func GetInt(name string) int {
	if value, err := strconv.Atoi(Get(name)); err == nil {
		return value
	}

	return 0
}
