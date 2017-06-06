package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/chakrit/go-bunyan"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

type testClock struct {
	time time.Time
}

func (t *testClock) Now() time.Time { return t.time }

func TestGetDaysSinceLastActive(t *testing.T) {

	var parseTests = []struct {
		lastActive string
		expected   int
		now        time.Time
	}{
		{"", -1, time.Now()},
		{"2016", -1, time.Now()},
		{"2006-01-02T15:04:05+00:00", 10, time.Date(2006, 01, 12, 23, 0, 0, 0, time.UTC)},
		{"2016-06-02T16:46:00+00:00", 0, time.Date(2016, 06, 01, 23, 0, 0, 0, time.UTC)},
	}

	for _, tt := range parseTests {

		mockStats := hipchat.RoomStatistics{
			LastActive: tt.lastActive,
		}

		mockJob := Job{
			Log:      bunyan.NewStdLogger("test", bunyan.NilSink()),
			JobID:    "jobId",
			TenantID: "work.TenantID",
			Clock:    &testClock{tt.now},
		}

		daysSincelastActive := mockJob.GetDaysSinceLastActive(1, &mockStats)
		if daysSincelastActive != tt.expected {
			t.Error(fmt.Sprintf("getDaysSinceLastActive calculation was wrong. Expected=%d Actual=%d", tt.expected, daysSincelastActive))
		}
	}
}

func TestGetDaysSinceCreated(t *testing.T) {

	var parseTests = []struct {
		created  string
		expected int
		now      time.Time
	}{
		{"", -1, time.Now()},
		{"2016", -1, time.Now()},
		{"2006-01-02T15:04:05+00:00", 10, time.Date(2006, 01, 12, 23, 0, 0, 0, time.UTC)},
		{"2016-06-02T16:46:00+00:00", 0, time.Date(2016, 06, 01, 23, 0, 0, 0, time.UTC)},
	}

	for _, tt := range parseTests {

		mockRoom := hipchat.Room{
			ID:      1,
			Created: tt.created,
		}

		mockJob := Job{
			Log:      bunyan.NewStdLogger("test", bunyan.NilSink()),
			JobID:    "jobId",
			TenantID: "work.TenantID",
			Clock:    &testClock{tt.now},
		}

		daysSincelastActive := mockJob.GetDaysSinceCreated(&mockRoom)
		if daysSincelastActive != tt.expected {
			t.Error(fmt.Sprintf("getDaysSinceLastActive calculation was wrong. Expected=%d Actual=%d", tt.expected, daysSincelastActive))
		}
	}
}

func TestShouldArchiveRoomTopic(t *testing.T) {
	var shouldArchiveTests = []struct {
		topic               string
		shouldArchive       bool
		daysSincelastActive int
	}{
		{"do not archive", false, 10},
		{"A valid topic | do not archive", false, 10},
		{"A valid topic | DO NOT ARCHIVE", false, 10},
		{"A valid topic with uñicode | DO NOT ARCHIVE", false, 10},
		{"archive now", true, 10},
		{"archive now", false, 1},
		{"", true, 10},
	}

	for _, tt := range shouldArchiveTests {

		mockJob := Job{
			Log:      bunyan.NewStdLogger("test", bunyan.NilSink()),
			JobID:    "jobId",
			TenantID: "work.TenantID",
		}

		shouldArchive := mockJob.ShouldArchiveRoom(1, tt.daysSincelastActive, 7, tt.topic)
		if shouldArchive != tt.shouldArchive {
			t.Error(fmt.Sprintf("ShouldArchiveRoom was wrong. Expected=%v Actual=%v Topic=%s", tt.shouldArchive, shouldArchive, tt.topic))
		}
	}
}
