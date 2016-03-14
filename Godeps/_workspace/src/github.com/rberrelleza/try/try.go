package try

import (
	"errors"
	"time"
	"math"
	"math/rand"
)

// MaxRetries is the maximum number of retries before bailing.
var MaxRetries = 10

var errMaxRetriesReached = errors.New("exceeded retry limit")

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// BackoffStrategy is used to determine how long a retry request should wait until attempted
type BackoffStrategy func(retry int) time.Duration

// Do keeps trying the function until the second argument
// returns false, or no error is returned.
func Do(fn Func) error {
	return DoWithBackoff(fn, DefaultBackoff)
}

func DoWithBackoff(fn Func, backoff BackoffStrategy) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > MaxRetries {
			return errMaxRetriesReached
		}

		// prevent a 0 from causing the tick to block, pass additional microsecond
		<-time.Tick(backoff(attempt) + 1*time.Microsecond)
	}
	return err
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}

// DefaultBackoff always returns 0 seconds
func DefaultBackoff(_ int) time.Duration {
	return 0 * time.Second
}

// ExponentialJitterBackoff returns ever increasing backoffs by a power of 2
// with +/- 0-33% to prevent sychronized reuqests.
func ExponentialJitterBackoff(i int) time.Duration {
	return jitter(int(math.Pow(2, float64(i))))
}

// jitter keeps the +/- 0-33% logic in one place
func jitter(i int) time.Duration {
	ms := i * 1000

	maxJitter := ms / 3

	rand.Seed(time.Now().Unix())
	jitter := rand.Intn(maxJitter + 1)

	if rand.Intn(2) == 1 {
		ms = ms + jitter
	} else {
		ms = ms - jitter
	}

	// a jitter of 0 messes up the time.Tick chan
	if ms <= 0 {
		ms = 1
	}

	return time.Duration(ms) * time.Millisecond
}
