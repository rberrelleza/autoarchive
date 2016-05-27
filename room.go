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

type Room struct {
	roomId      int
	last_active string
}

type options struct {
	StartIndex int `url:"start-index"`
	MaxResults int `url:"max-results"`
}

func (w *Worker) GetRooms(client *hipchat.Client) ([]hipchat.Room, error) {
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
			rooms, response, err = client.Room.List(opt)
			return attempt < 5, err // try 5 times
		}, try.ExponentialJitterBackoff)

		if err != nil {
			w.Log.Errorf("Client.CreateClient returns an error %v", response)
			break
		}

		roomList = append(roomList, rooms.Items...)

		if rooms.Links.Next == "" {
			w.Log.Debugf("client.Room.List retreieved all the rooms")
			break
		} else {
			startIndex += maxResults
		}
	}

	return roomList, err
}

func (w *Worker) MaybeArchiveRoom(tenantID string, roomId int, threshold int, client *hipchat.Client) {
	daysSinceLastActive := w.getDaysSinceLastActive(roomId, client)
	remainingIdleDaysAllowed := daysSinceLastActive - threshold

	if remainingIdleDaysAllowed >= 0 {
		w.archiveRoom(tenantID, roomId, client, daysSinceLastActive)
	}
}

func (w *Worker) getDaysSinceLastActive(roomID int, client *hipchat.Client) int {
	var response *http.Response
	var stats *hipchat.RoomStatistics

	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		stats, response, err = client.Room.GetStatistics(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

	if err != nil {
		w.Log.Debugf("Client.Room.GetStatistics returns an error %v", response)

	} else {
		if stats.LastActive == "" {
			w.Log.Infof("last_active is empty for rid-%d %s", roomID, stats.LastActive)
		} else {
			w.Log.Debugf("rid-%d last_active %v", roomID, stats.LastActive)

			lastActive, err := time.Parse(timeFormat, stats.LastActive)
			if err != nil {
				w.Log.Debugf("Couldn't parse rid-%d date error: %v", roomID, err)
			} else {
				delta := time.Now().Sub(lastActive)
				deltaInDays := int(delta.Hours() / 24) //assumes every day has 24 hours, not DST aware
				w.Log.Debugf("rid-%d has been idle for %d days", roomID, deltaInDays)
				return deltaInDays
			}
		}
	}

	// default case if the room doesn't have an last_active date or
	// if there was an error
	return 0
}

func (w *Worker) archiveRoom(tenantID string, roomID int, client *hipchat.Client, idleDays int) {
	var response *http.Response
	var room *hipchat.Room

	err := try.DoWithBackoff(func(attempt int) (bool, error) {
		var err error
		room, response, err = client.Room.Get(strconv.Itoa(roomID))
		return attempt < 5, err // try 5 times
	}, try.ExponentialJitterBackoff)

	if err != nil {
		w.Log.Errorf("Client.Room.Get returned an error %v", response)
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

	message := fmt.Sprintf("@%s This room was archived since it has been inactive for %d days. Go to https://hipchat.com/rooms/archive/%d to unarchive it.", room.Owner.MentionName, idleDays, roomID)

	dryRun := util.Env.GetInt("DRYRUN_ENV")
	if dryRun == 1 {
		w.Log.Infof("Would've archived tid-%s rid-%d: %s", tenantID, roomID, message)
	} else {
		w.notifyArchival(roomID, message, client)

		resp, err := client.Room.Update(strconv.Itoa(roomID), &updateRequest)

		if err != nil {
			w.Log.Errorf("Client.Room.Update returned an error when archiving")
			contents, err := ioutil.ReadAll(resp.Body)
			w.Log.Errorf("%s %s", contents, err)
		} else {
			w.Log.Infof("Archived tid-%d rid-%d", tenantID, roomID)
		}
	}
}

func (w *Worker) notifyArchival(roomID int, message string, client *hipchat.Client) {
	notificationRequest := hipchat.NotificationRequest{
		Message:       message,
		Notify:        true,
		MessageFormat: "text",
	}

	resp, err := client.Room.Notification(strconv.Itoa(roomID), &notificationRequest)

	if err != nil {
		w.Log.Errorf("Client.Room.Notification returned an error when archiving %v", resp)
		contents, err := ioutil.ReadAll(resp.Body)
		w.Log.Errorf("%s %s", contents, err)
	}
}
