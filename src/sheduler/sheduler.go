package sheduler

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

type Sheduler struct {
	Cron cron.Cron
}

func NewSheduler(location string) Sheduler {
	locationTime, _ := time.LoadLocation(location)
	cron := cron.New(cron.WithLocation(locationTime))
	sheduler := Sheduler{Cron: *cron}
	go sheduler.Cron.Run()
	return sheduler
}

func (s *Sheduler) AddTask(spec string, cmd func()) (cron.EntryID, error) {
	entryID, err := s.Cron.AddFunc(spec, cmd)
	log.Println(fmt.Sprintf("Task %v created at %s", entryID, spec))
	return entryID, err
}

func (s *Sheduler) RemoveTask(entryID cron.EntryID) {
	log.Println(fmt.Sprintf("Task %v removed", entryID))
	s.Cron.Remove(entryID)
}
