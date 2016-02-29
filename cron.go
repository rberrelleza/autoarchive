package main

import (
  "github.com/jasonlvhit/gocron"
  "os"
  "time"
  )

var WorkQueue = make(chan WorkRequest, 100)

func auto_archive() {
  groups, err := GetAllGroups()
  checkErr(err)

  log.Debugf("Found %d groups", len(groups))

  for i := range groups {
    log.Debugf("Start archiving group %d", groups[i].groupId)
    work := WorkRequest{gid: groups[i].groupId}
    WorkQueue <- work
  }
}

func StartCron() {
  go func() {
    duration, error := time.ParseDuration(os.Getenv("WAIT_CHECKS"))
    log.Infof("Archiver will run every %s", duration)
    checkErr(error)
    seconds := uint64(duration.Seconds())
    gocron.Every(seconds).Seconds().Do(auto_archive)
    <- gocron.Start()
  }()
}
