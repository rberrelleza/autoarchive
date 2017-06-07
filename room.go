package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

const (
	// See http://golang.org/pkg/time/#Parse
	timeFormat = "2006-01-02T15:04:05+00:00"
)

const (
	topic = "do not archive"
)

// GetRooms retrieves all the active rooms for a specific tenant. This function calls the HipChat /room API, batching
// results in groups of 1000.
// It returns a list of rooms, and or any errors.
func (j *Job) GetRooms() ([]hipchat.Room, error) {
	var roomList []hipchat.Room
	var response *http.Response
	var err error
	startIndex := 0
	maxResults := 1000

	for {
		var rooms *hipchat.Rooms
		opt := &hipchat.RoomsListOptions{
			ListOptions:     hipchat.ListOptions{StartIndex: startIndex, MaxResults: maxResults},
			ExpandOptions:   hipchat.ExpandOptions{Expand: "items"},
			IncludePrivate:  true,
			IncludeArchived: false}

		rooms, response, err = j.Client.Room.List(opt)

		if err != nil {
			j.Log.Errorf("Client.CreateClient returns an error %v", response)
			break
		}

		roomList = append(roomList, rooms.Items...)

		if rooms.Links.Next == "" {
			j.Log.Debugf("client.Room.List retreieved all the rooms")
			break
		} else {
			startIndex += maxResults
		}
	}

	j.Log.Infof("Retrieved %d rooms", len(roomList))
	return roomList, err
}

// ShouldArchiveRoom returns true if a rooom has been inactive for longer than the threshold
func (j *Job) ShouldArchiveRoom(roomID, daysSinceLastActive, threshold int, roomTopic string) bool {
	shouldArchive := false

	s := strings.ToLower(roomTopic)

	if strings.Contains(s, topic) {
		j.Log.Record("rid", roomID).Infof("Skipping due to topic overwrite")
	} else {
		remainingIdleDaysAllowed := daysSinceLastActive - threshold
		if remainingIdleDaysAllowed >= 0 {
			shouldArchive = true
		}
	}

	return shouldArchive
}

// TouchRoom sends a message to the room, so the last_active date won't be empty the next time the autoarchiver runs
func (j *Job) TouchRoom(roomID int, threshold int) {
	if j.DryRun {
		j.Log.Record("rid", roomID).Infof("Would've updated last_active")
	} else {
		message := fmt.Sprintf("This room hasn't been used in a while, but I can't tell how long (okay).  The room will be archived if it remains inactive for the next %d days.", threshold)
		j.notify(roomID, message)
	}
}

// GetDaysSinceLastActive calculates how many days  has a room been inactive,
// based on the current time and the value return from the Room Stats hipchat API
func (j *Job) GetDaysSinceLastActive(roomID int, stats *hipchat.RoomStatistics) int {
	var deltaInDays = -1

	if stats.LastActive == "" {
		j.Log.Record("rid", roomID).Debugf("last_active is empty")
	} else {
		j.Log.Record("rid", roomID).Debugf("last_active %v", stats.LastActive)
		lastActive, err := time.Parse(timeFormat, stats.LastActive)
		if err != nil {
			j.Log.Record("rid", roomID).Errorf("Couldn't parse date error: %v", err)
		} else {
			delta := j.Clock.Now().Sub(lastActive)
			j.Log.Record("rid", roomID).Debugf("Has been idle for %s", delta)
			deltaInDays = int(delta.Hours() / 24) //assumes every day has 24 hours, not DST aware
			j.Log.Record("rid", roomID).Debugf("Has been idle for %d days", deltaInDays)
		}
	}

	return deltaInDays
}

// GetDaysSinceCreated calculates how many days since the room was created
// based on the current time and the value return from the Room hipchat API
func (j *Job) GetDaysSinceCreated(room *hipchat.Room) int {
	var deltaInDays = -1
	created, err := time.Parse(timeFormat, room.Created)
	if err != nil {
		j.Log.Record("rid", room.ID).Errorf("Couldn't parse date error: %v", err)
	} else {
		// TODO pass the JobID and the Time, so we don't deal with objects here
		delta := j.Clock.Now().Sub(created)
		deltaInDays = int(delta.Hours() / 24) //assumes every day has 24 hours, not DST aware
		j.Log.Record("rid", room.ID).Debugf("Was created and idle for %d days", deltaInDays)
	}

	return deltaInDays
}

// GetRoomStats queries the hipchat api to get the RoomStatistics of the roomID
func (j *Job) GetRoomStats(roomID int) (*hipchat.RoomStatistics, error) {
	stats, _, err := j.Client.Room.GetStatistics(strconv.Itoa(roomID))
	return stats, err
}

// ArchiveRoom calls the hipchat API to archive the room. It sends a message while archiving so the owner of the room will know what
// happened to her room.
func (j *Job) ArchiveRoom(roomID int, daysSinceLastActive int) error {
	var response *http.Response
	var room *hipchat.Room

	room, err := j.GetRoom(roomID)

	if err != nil {
		j.Log.Record("rid", roomID).Errorf("Client.Room.Get returned an error %v", response)
		return err
	}

	room.IsArchived = true
	ownerID := hipchat.ID{ID: strconv.Itoa(room.Owner.ID)}
	updateRequest := hipchat.UpdateRoomRequest{
		Name:          room.Name,
		Topic:         room.Topic,
		IsGuestAccess: room.IsGuestAccessible,
		IsArchived:    true,
		Privacy:       room.Privacy,
		Owner:         ownerID,
	}

	message := fmt.Sprintf("Archiving the room since it has been inactive for %d days. Go to %s/rooms/archive/%d to unarchive it.", daysSinceLastActive, j.HipChatURL, roomID)

	err = nil
	if j.DryRun {
		j.Log.Record("rid", roomID).Infof("Would've archived: %s", message)
	} else {
		j.notify(roomID, message)
		resp, err := j.Client.Room.Update(strconv.Itoa(roomID), &updateRequest)

		if err != nil {
			j.Log.Record("rid", roomID).Errorf("Client.Room.Update returned an error when archiving")
			contents, _ := ioutil.ReadAll(resp.Body)
			j.Log.Record("rid", roomID).Errorf("%s %s", contents, err)
		} else {
			j.Log.Record("rid", roomID).Infof("Archived room, idle for %d days", daysSinceLastActive)
		}
	}

	return err
}

// GetRoom calls the hipchat api to get the full room object
func (j *Job) GetRoom(roomID int) (*hipchat.Room, error) {
	room, _, err := j.Client.Room.Get(strconv.Itoa(roomID))
	return room, err
}

func (j *Job) notify(roomID int, message string) {
	notificationRequest := hipchat.NotificationRequest{
		Message:       message,
		Notify:        true,
		MessageFormat: "text",
	}

	resp, err := j.Client.Room.Notification(strconv.Itoa(roomID), &notificationRequest)

	if err != nil {
		j.Log.Errorf("Client.Room.Notification returned an error when archiving %v", resp)
		contents, err := ioutil.ReadAll(resp.Body)
		j.Log.Errorf("%s %s", contents, err)
	}
}
