package main

import (
  "github.com/jasonlvhit/gocron"
  "time"
  )

var WorkQueue = make(chan WorkRequest, 100)

func auto_archive(context *Context) {
  groups, err := GetAllGroups(context)
  checkErr(err)

  log.Debugf("Found %d groups", len(groups))

  for i := range groups {
    log.Debugf("Start archiving group %d", groups[i].groupId)
    work := WorkRequest{gid: groups[i].groupId}
    WorkQueue <- work
  }
}

func StartCron(context *Context, schedule time.Duration) {
  go func() {
    seconds := uint64(schedule.Seconds())
    log.Infof("Archiver will run every %s", schedule)
    gocron.Every(seconds).Seconds().Do(auto_archive, context)
    <- gocron.Start()
  }()
}
