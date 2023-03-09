package sheduler

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func (s *Sheduler) ReinderBackUpDataBase(dataBaseName string) {
	arg := fmt.Sprintf("--dbname=postgresql://%s:%s@%s:5432/%s", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), dataBaseName)
	cmnd := exec.Command(os.Getenv("PG_DIR")+"pg_basebackup", arg)

	log.Println(cmnd)
	err := cmnd.Start()
	if err != nil {
		log.Println(err)
	}
}

func (s *Sheduler) CreateBackUpDataBase(dataBaseName string) {
	arg := fmt.Sprintf("--dbname=postgresql://%s:%s@%s:5432/* -D \"C:\\Backup\"", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"))
	cmnd := exec.Command(fmt.Sprintf("%s\\%s", os.Getenv("PG_DIR"), "pg_basebackup"), arg)

	log.Println(cmnd)
	err := cmnd.Start()
	if err != nil {
		log.Println(err)
	}
}
