package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type DataBase struct {
	Name  string
	Tasks []Task
}

type Time struct {
	H int
	M int
	S int
}

func Itoa(i int) string {
	r := strconv.Itoa(i)
	if r == "0" {
		return "00"
	}
	return r
}

func (t Time) ToString() string {
	return fmt.Sprintf("%s:%s:%s", Itoa(t.H), Itoa(t.M), Itoa(t.S))
}

func CheckTime(t string, p string) (error, int) {
	if len(t) != 2 {
		return errors.New("empty name"), 0
	}
	i, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("Failed to convert to a number"), 0
	}
	switch p {
	case "h":
		if i > 23 {
			return errors.New("More than 12"), 0
		}
	case "m", "s":
		if i > 60 {
			return errors.New("More than 60"), 0
		}
	}
	return nil, i
}

type Task struct {
	Name     string
	TaskType string
	Period   string
	Time     Time
}

func Save(data []DataBase) {
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	filename, err := os.Create("databases.json")
	defer filename.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config saved")
	filename.Write(json_data)
}

func Load() []DataBase {

	data, err := ioutil.ReadFile("databases.json")
	if err != nil {
		log.Println(err)
	}
	var DataBasesFromFile []DataBase
	json.Unmarshal(data, &DataBasesFromFile)

	log.Println(DataBasesFromFile)

	return DataBasesFromFile
}
