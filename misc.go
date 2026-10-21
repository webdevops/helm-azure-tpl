package main

import (
	"os"
)

func setOsEnvIfUnset(name, value string) {
	if envVal := os.Getenv(name); envVal == "" {
		if err := os.Setenv(name, value); err != nil {
			panic(err)
		}
	}
}
