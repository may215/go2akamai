package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

/* Return the configuration data from the config file */
func getConfig(config_file string, config_instance interface{}) *errorHandler {
	if config_file == "" {
		return &errorHandler{errors.New("Need to provide config file"), "Need to provide config file", 100001}
	}

	file, err := os.Open("./" + config_file + ".conf")
	if err != nil {
		return &errorHandler{err, err.Error(), 100002}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var conf = strings.Join(lines, " ")
	b := []byte(conf)
	m_err := json.Unmarshal(b, &config_instance)
	if m_err != nil {
		return &errorHandler{m_err, m_err.Error(), 100003}
	}

	return &errorHandler{err, "success reading config file", 0}
}
