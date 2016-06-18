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

		daysSincelastActive := mockJob.getDaysSinceLastActive(1, &mockStats)
		if daysSincelastActive != tt.expected {
			t.Error(fmt.Sprintf("getDaysSinceLastActive calculation was wrong. Expected=%d Actual=%d", tt.expected, daysSincelastActive))
		}
	}
}
