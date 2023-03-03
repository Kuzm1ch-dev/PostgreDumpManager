package common

import (
	"PostgresDumpManager/src/sheduler"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
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
}

func Itoa(i int) string {
	r := strconv.Itoa(i)
	if r == "0" {
		return "00"
	}
	return r
}

func (t Time) ToString() string {
	return fmt.Sprintf("%s:%s", Itoa(t.H), Itoa(t.M))
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
	case "m":
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
	EntryID  cron.EntryID
}

func Save(data []DataBase, sheduler sheduler.Sheduler) {
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
	sheduler.Cron.Stop()
	go sheduler.Cron.Start()
}

func Load(sheduler sheduler.Sheduler) []DataBase {

	data, err := ioutil.ReadFile("databases.json")
	if err != nil {
		log.Println(err)
	}
	var DataBasesFromFile []DataBase
	json.Unmarshal(data, &DataBasesFromFile)

	log.Println(DataBasesFromFile)

	for _, database := range DataBasesFromFile {
		for _, task := range database.Tasks {
			entryID, err := sheduler.AddTask(fmt.Sprintf("%d %d * * 0-6", task.Time.M, task.Time.H), func() { sheduler.CreateBackUpDataBase("Hello") })
			if err != nil {
				log.Println(err)
			}
			task.EntryID = entryID
		}
	}

	return DataBasesFromFile
}
