package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Host          string
	User          string
	Password      string
	PostgreBinDir string
}

func (conf *Config) SaveConfigInFile() {
	os.Setenv("PG_HOST", conf.Host)
	os.Setenv("PG_USER", conf.User)
	os.Setenv("PG_PASS", conf.Password)
	os.Setenv("PG_DIR", conf.PostgreBinDir)
	log.Println(conf)
	json_data, err := json.Marshal(conf)
	if err != nil {
		log.Println(err)
	}
	filename, err := os.Create("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config saved")
	filename.Write(json_data)
	filename.Close()
}

func (conf *Config) LoadConfigFromFile() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println(err)
		file, err := os.Create("config.json")
		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		file.Close()
		return
	}
	var config Config
	json.Unmarshal(data, &config)
	conf = &config
	log.Println(conf)
	os.Setenv("PG_HOST", conf.Host)
	os.Setenv("PG_USER", conf.User)
	os.Setenv("PG_PASS", conf.Password)
	os.Setenv("PG_DIR", conf.PostgreBinDir)
	log.Println("Config loaded")
}
