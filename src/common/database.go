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

type DataBaseStorage struct {
	Storage []DataBase
}

func (dbs *DataBaseStorage) CreateDataBase(name string) {
	dbs.Storage = append(dbs.Storage, DataBase{name, []Task{}})
}

func (dbs *DataBaseStorage) RemoveDataBase(index int) {
	dbs.Storage = append(dbs.Storage[:index], dbs.Storage[index+1:]...)
}

func (dbs *DataBaseStorage) GetDataBase(index int) *DataBase {
	return &dbs.Storage[index]
}

type DataBase struct {
	Name  string
	Tasks []Task
}

func (db *DataBase) CreateTask(name string, entryID cron.EntryID) {
	db.Tasks = append(db.Tasks, Task{name, "Backup", "Every day", Time{12, 0}, entryID})
}

func (db *DataBase) RemoveTask(index int) {
	db.Tasks = append(db.Tasks[:index], db.Tasks[index+1:]...)
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

func (dbs *DataBaseStorage) SaveDataBaseInFile(sheduler sheduler.Sheduler) {
	json_data, err := json.Marshal(dbs.Storage)
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

func (dbs *DataBaseStorage) LoadDataBaseFromFile(sheduler sheduler.Sheduler) {
	data, err := ioutil.ReadFile("databases.json")
	if err != nil {
		log.Println(err)
		file, err := os.Create("databases.json")
		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		file.Close()
		return
	}
	var DataBasesFromFile []DataBase
	json.Unmarshal(data, &DataBasesFromFile)

	for _, database := range DataBasesFromFile {
		for _, task := range database.Tasks {
			entryID, err := sheduler.AddTask(fmt.Sprintf("%d %d * * 0-6", task.Time.M, task.Time.H), func() { sheduler.CreateBackUpDataBase("Hello") })
			if err != nil {
				log.Println(err)
			}
			task.EntryID = entryID
		}
	}

	dbs.Storage = DataBasesFromFile
}
