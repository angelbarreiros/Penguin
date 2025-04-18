package main

import (
	cronengine "angelotero/commonBackend/cronEngine"
	"log"
	"os"
	"time"
)

func main() {
	var file, err = os.OpenFile("./logs/report.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	var scheduler = cronengine.StartScheduler()
	var id, errJ = scheduler.ScheduleJob("test", time.Now().Add(10*time.Second))
	if errJ != nil {
		panic(errJ)
	}
	log.Println(id)
	scheduler.RemoveJob(id)

	time.Sleep(40 * time.Second)

}
