package main

import (
	"github.com/asdine/storm"
	"github.com/joaquinpf/readordie/internal/app/readordie"
	"log"
)

func main() {
	readordie.InitLogging()

	stormdb, err := storm.Open("readordie.db")
	if err != nil {
		log.Panic(err)
	}
	defer stormdb.Close()

	readordie.NewCronEnv(stormdb).Start(30)
	readordie.NewServerEnv(stormdb).Start()
}
