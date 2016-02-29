package main

import (
  "github.com/tbruyelle/hipchat-go/hipchat"
  "strconv"
  "time"
  "os"
  "fmt"
  "io/ioutil"
)

const (
    // See http://golang.org/pkg/time/#Parse
    timeFormat = "2006-01-02T15:04:05+00:00"

    // room will be archive if idle for more than 90 days
    maxRoomIdleness = 90
)

var dryRun, _ =  strconv.ParseBool(os.Getenv("DRY_RUN"))

type Room struct {
	roomId  int
	last_active string
}

func getRooms(groupId int, client *hipchat.Client) ([]hipchat.Room) {
  rooms, response, err := client.Room.List()
  if err != nil {
    log.Errorf("Client.CreateClient returns an error %v", response)
  }

  return rooms.Items
}

func maybeArchiveRoom(groupId int, roomId int, client *hipchat.Client) {
  daysSinceLastActive := getDaysSinceLastActive(roomId, client)

  remainingIdleDaysAllowed := daysSinceLastActive - maxRoomIdleness

  if remainingIdleDaysAllowed >= 0 {
    archiveRoom(groupId, roomId, client, daysSinceLastActive)
  }
}

func getDaysSinceLastActive(roomId int, client *hipchat.Client) (int) {
  stats, response, err := client.Room.GetStatistics(strconv.Itoa(roomId))

  if err != nil {
    log.Debugf("Client.Room.GetStatistics returns an error %v", response)

  } else {
    if stats.LastActive == "" {
        log.Debugf("last_active is empty for rid-%d %s", roomId, stats.LastActive)
    } else {
      log.Debugf("rid-%d last_active %v", roomId, stats.LastActive)

      lastActive, err := time.Parse(timeFormat, stats.LastActive)
      if err != nil {
          log.Debugf("Couldn't parse rid-%d date error: %v", roomId, err)
      } else {
        delta := time.Now().Sub(lastActive)
        deltaInDays := int(delta.Hours()/24) //assumes every day has 24 hours, not DST aware
        log.Debugf("rid-%d has been idle for %d days", roomId, deltaInDays)
        return deltaInDays
      }
    }
  }

  // default case if the room doesn't have an last_active date or
  // if there was an error
  return 0
}

func archiveRoom(groupId int, roomId int, client *hipchat.Client, idleDays int){
  room, response, err := client.Room.Get(strconv.Itoa(roomId))
  if err != nil {
    log.Errorf("Client.Room.Get returned an error %v", response)
    return
  }

  room.IsArchived = true
  owner_id := hipchat.ID { ID: strconv.Itoa(room.Owner.ID) }
  updateRequest := hipchat.UpdateRoomRequest {
    Name: room.Name,
    Topic: room.Topic,
    IsGuestAccess: room.IsGuestAccessible,
    IsArchived: true,
    Privacy: room.Privacy,
    Owner: owner_id,
  }

  message := fmt.Sprintf("@%s This room was archived since it has been inactive for %d days. Go to https://hipchat.com/rooms/archive/%d to unarchive it.", room.Owner.MentionName, idleDays, roomId)


  if dryRun {
    log.Infof("Would've archived gid-%d rid-%d: %s", groupId, roomId, message)
  } else {

    notifyArchival(roomId, message, client)

    resp, err := client.Room.Update(strconv.Itoa(roomId), &updateRequest)

    if err != nil {
      log.Errorf("Client.Room.Update returned an error when archiving")
      contents, err := ioutil.ReadAll(resp.Body)
      log.Errorf("%s %s", contents, err)
    } else {
      log.Infof("Archived gid-%d rid-%d", groupId, roomId)
    }
  }
}

func notifyArchival(roomId int, message string, client *hipchat.Client){
  notificationRequest := hipchat.NotificationRequest {
    Message: message,
    Notify: true,
    MessageFormat: "text",
  }

  resp, err := client.Room.Notification(strconv.Itoa(roomId), &notificationRequest)
  if err != nil {
    log.Errorf("Client.Room.Notification returned an error when archiving %v", resp)
    contents, err := ioutil.ReadAll(resp.Body)
    log.Errorf("%s %s", contents, err)
  }
}
