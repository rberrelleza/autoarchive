package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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

	stats, err := j.getRoomStats(roomID)

	if err != nil {
		j.Log.Record("rid", roomID).Errorf("Client.Room.GetStatistics returned an error %v", err)
	} else {
		daysSinceLastActive := j.getDaysSinceLastActive(roomID, stats)

		if daysSinceLastActive == -1 {
			if j.DryRun {
				j.Log.Record("rid", roomID).Infof("Would've updated last_active")
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
	}

	return false
}

func (j *Job) getDaysSinceLastActive(roomID int, stats *hipchat.RoomStatistics) int {
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

	// default case if the room doesn't have an last_active date or
	// if there was an error
	return deltaInDays
}

func (j *Job) getRoomStats(roomID int) (*hipchat.RoomStatistics, error) {
	var stats *hipchat.RoomStatistics
	var response *http.Response

	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		stats, response, err = j.Client.Room.GetStatistics(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

	return stats, err
}

func (j *Job) archiveRoom(roomID int, idleDays int) {
	var response *http.Response
	var room *hipchat.Room

	room, err := j.getRoom(roomID)

	if err != nil {
		j.Log.Record("rid", roomID).Errorf("Client.Room.Get returned an error %v", response)
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

	message := fmt.Sprintf("Archiving the room since it has been inactive for %d days. Go to %s/rooms/archive/%d to unarchive it.", idleDays, j.HipChatURL, roomID)

	if j.DryRun {
		j.Log.Record("rid", roomID).Infof("Would've archived: %s", message)
	} else {
		j.notify(roomID, message)
		resp, err := j.Client.Room.Update(strconv.Itoa(roomID), &updateRequest)

		if err != nil {
			j.Log.Record("rid", roomID).Errorf("Client.Room.Update returned an error when archiving")
			contents, err := ioutil.ReadAll(resp.Body)
			j.Log.Record("rid", roomID).Errorf("%s %s", contents, err)
		} else {
			j.Log.Record("rid", roomID).Infof("Archived", roomID)
		}
	}
}

func (j *Job) getRoom(roomID int) (*hipchat.Room, error) {
	var room *hipchat.Room
	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		room, _, err = j.Client.Room.Get(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

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
