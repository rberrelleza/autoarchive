package main

import (
	"github.com/jasonlvhit/gocron"
	"time"
)

var WorkQueue = make(chan WorkRequest, 100)

func (context *Context) RunScheduler(schedule *string) {
	log.Infof("Starting the scheduler")
	duration, error := time.ParseDuration(*schedule)
	checkErr(error)

	go func() {
		seconds := uint64(duration.Seconds())
		log.Infof("Archiver will run every %s", *schedule)
		gocron.Every(seconds).Seconds().Do(autoArchive, context)
		<-gocron.Start()
	}()

}

func autoArchive(context *Context) {
	groups, err := GetAllGroups(context)
	checkErr(err)

	log.Debugf("Found %d groups", len(groups))

	for i := range groups {
		log.Debugf("Start archiving group %d", groups[i].groupId)
		work := WorkRequest{gid: groups[i].groupId}
		WorkQueue <- work
	}
}
