package sheduler

import (
	"log"
	"os"
	"os/exec"
)

func (s *Sheduler) ReinderBackUpDataBase(dataBaseName string) {
	cmnd := exec.Command(os.Getenv("PG_DIR")+"pg_dump", "")
	err := cmnd.Start()
	if err != nil {
		log.Println(err)
	}
}

func (s *Sheduler) CreateBackUpDataBase(dataBaseName string) {
	cmnd := exec.Command(os.Getenv("PG_DIR")+"reindexdb", "")
	err := cmnd.Start()
	if err != nil {
		log.Println(err)
	}
}
