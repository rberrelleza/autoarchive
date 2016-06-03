package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"github.com/rberrelleza/try"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

const (
	// See http://golang.org/pkg/time/#Parse
	timeFormat = "2006-01-02T15:04:05+00:00"
)

type options struct {
	StartIndex int `url:"start-index"`
	MaxResults int `url:"max-results"`
}

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
		opt := options{startIndex, maxResults}

		err = try.DoWithBackoff(func(attempt int) (bool, error) {
			var err error
			rooms, response, err = j.Client.Room.List(opt)
			return attempt < 5, err // try 5 times
		}, try.ExponentialJitterBackoff)

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

	return roomList, err
}

// MaybeArchiveRoom retrieves the last active date of a room, compares that to the threshold passed, and archives the room if the result is negative.
// The function will only 'pretend' to archive if the DRYRUN_ENV env var is set.
// If a room doesn't have a last active date, the function will send a message to said room, to initialize that date.
func (j *Job) MaybeArchiveRoom(roomID int, threshold int) bool {
	daysSinceLastActive := j.getDaysSinceLastActive(roomID)

	if daysSinceLastActive == -1 {
		if isDryRun() {
			j.Log.Infof("Would've updated last_active of rid-%d", roomID)
		} else {
			message := fmt.Sprintf("This room hasn't been used in a while, but I can't tell how long (okay).  The room will be archived if it remains inactive for the next %d days.", threshold)
			j.notify(roomID, message)
		}
	} else {

		remainingIdleDaysAllowed := daysSinceLastActive - threshold

		if remainingIdleDaysAllowed >= 0 {
			j.archiveRoom(roomID, daysSinceLastActive)
			return true
		}
	}

	return false
}

func (j *Job) getDaysSinceLastActive(roomID int) int {
	var response *http.Response
	var stats *hipchat.RoomStatistics

	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		stats, response, err = j.Client.Room.GetStatistics(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

	var deltaInDays int

	if err != nil {
		j.Log.Errorf("Client.Room.GetStatistics returns an error %v", response)
	} else {
		if stats.LastActive == "" {
			j.Log.Debugf("last_active is empty for rid-%d %s", roomID, stats.LastActive)
			deltaInDays = -1
		} else {
			j.Log.Debugf("rid-%d last_active %v", roomID, stats.LastActive)

			lastActive, err := time.Parse(timeFormat, stats.LastActive)
			if err != nil {
				j.Log.Errorf("Couldn't parse rid-%d date error: %v", roomID, err)
			} else {
				delta := time.Now().Sub(lastActive)
				deltaInDays = int(delta.Hours() / 24) //assumes every day has 24 hours, not DST aware
				j.Log.Debugf("rid-%d has been idle for %d days", roomID, deltaInDays)
			}
		}
	}

	// default case if the room doesn't have an last_active date or
	// if there was an error
	return deltaInDays
}

func (j *Job) archiveRoom(roomID int, idleDays int) {
	var response *http.Response
	var room *hipchat.Room

	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		room, response, err = j.Client.Room.Get(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

	if err != nil {
		j.Log.Errorf("Client.Room.Get rid-%d returned an error %v", roomID, response)
		return
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

	message := fmt.Sprintf("Archiving the room since it has been inactive for %d days. Go to https://hipchat.com/rooms/archive/%d to unarchive it.", idleDays, roomID)

	if isDryRun() {
		j.Log.Infof("Would've archived rid-%d: %s", roomID, message)
	} else {
		j.notify(roomID, message)
		resp, err := j.Client.Room.Update(strconv.Itoa(roomID), &updateRequest)

		if err != nil {
			j.Log.Errorf("Client.Room.Update returned an error when archiving")
			contents, err := ioutil.ReadAll(resp.Body)
			j.Log.Errorf("%s %s", contents, err)
		} else {
			j.Log.Infof("Archived rid-%d", roomID)
		}
	}
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

func isDryRun() bool {
	dryRun := util.Env.GetInt("DRYRUN_ENV")
	if dryRun == 1 {
		return true
	}

	return false
}
