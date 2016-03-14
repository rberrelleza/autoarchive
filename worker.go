package main

import (
	"github.com/tbruyelle/hipchat-go/hipchat"
)

var WorkerQueue chan chan WorkRequest

func StartWorkers() {
	b := NewBackendServer("hiparchiver.workers")
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, 4)

	// Now, create all of our workers.
	for i := 0; i < 4; i++ {
		b.Log.Infof("Starting worker-%d", i+1)
		worker := b.newWorker(i+1, WorkerQueue)
		worker.start(b)
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				b.Log.Debugf("Received work request")
				go func() {
					worker := <-WorkerQueue
					b.Log.Debugf("Dispatching work request")
					worker <- work
				}()
			}
		}
	}()
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func (s *Server) newWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		Log:         s.Log,
	}

	return worker
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w Worker) start(s *Server) {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				w.Log.Debugf("worker%d: Received work request for tid-%s", w.ID, work.TenantID)

				tenants := s.NewTenants()
				tenant, err := tenants.Get(work.TenantID)
				if err != nil {
					w.Log.Errorf("Coudn't find tid-%s", work.TenantID)
					return
				}

				tenantConfigurations := s.NewTenantConfigurations()
				tenantConfiguration, err := tenantConfigurations.Get(tenant.ID)

				if err != nil {
					w.Log.Errorf("Coudn't find a configuration for tid-%s", work.TenantID)
					return
				}

				credentials := hipchat.ClientCredentials{
					ClientID:     tenant.ID,
					ClientSecret: tenant.Secret,
				}

				newClient := hipchat.NewClient("")
				token, _, err := newClient.GenerateToken(
					credentials,
					[]string{hipchat.ScopeManageRooms, hipchat.ScopeViewGroup, hipchat.ScopeSendNotification, hipchat.ScopeAdminRoom})

				if err != nil {
					// this typically means the group uninstalled the plugin
					w.Log.Errorf("Client.GetAccessToken returns an error %v", err)
				} else {
					client := token.CreateClient()
					rooms, error := w.GetRooms(client)
					if error != nil {
						w.Log.Errorf("Failed to retrieve rooms for tid-%d", work.TenantID)
					} else {
						for _, room := range rooms {
							w.MaybeArchiveRoom(work.TenantID, room.ID, tenantConfiguration.Threshold, client)
						}
					}
				}
			case <-w.QuitChan:
				// We have been asked to stop.
				w.Log.Debugf("worker%d stopping\n", w.ID)
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
