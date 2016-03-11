package main

import (
	"github.com/tbruyelle/hipchat-go/hipchat"
)

var WorkerQueue chan chan WorkRequest

func (context *Context) RunDispatcher() {
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, context.nworkers)

	// Now, create all of our workers.
	for i := 0; i < context.nworkers; i++ {
		log.Infof("Starting worker-%d", i+1)
		worker := newWorker(i+1, WorkerQueue)
		worker.start(context)
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				log.Debug("Received work request")
				go func() {
					worker := <-WorkerQueue

					log.Debug("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func newWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w Worker) start(context *Context) {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				log.Debugf("worker%d: Received work request", w.ID)

				group, error := GetGroup(context, work.gid)
				checkErr(error)
				credentials := hipchat.ClientCredentials{
					ClientID:     group.oauthId,
					ClientSecret: group.oauthSecret,
				}

				newClient := hipchat.NewClient("")
				token, _, err := newClient.GenerateToken(
					credentials,
					[]string{hipchat.ScopeManageRooms, hipchat.ScopeViewGroup, hipchat.ScopeSendNotification, hipchat.ScopeAdminRoom})

				if err != nil {
					// this typically means the group uninstalled the plugin
					log.Errorf("Client.GetAccessToken returns an error %v", err)
				} else {
					client := token.CreateClient()
					rooms := getRooms(work.gid, client)
					for _, room := range rooms {
						maybeArchiveRoom(work.gid, room.ID, client)
					}
				}
			case <-w.QuitChan:
				// We have been asked to stop.
				log.Debugf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w Worker) stop() {
	go func() {
		w.QuitChan <- true
	}()
}
