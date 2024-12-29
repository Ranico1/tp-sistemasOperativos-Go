package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func CargarConfig[T any](filePath string) *T {
	var config *T

	configFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}
