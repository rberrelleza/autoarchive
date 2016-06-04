package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/satori/go.uuid"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

var WorkerQueue chan chan WorkRequest

// StartWorker starts the worker jobs of the process. It will start a number of workers equal to the value of the WORKERS_ENV env var, or 1
func StartWorker() {
	b := NewBackendServer("hiparchiver.workers")
	numWorkers := util.Env.GetIntOr("WORKERS_ENV", 1)
	maxRoomsToProcess := util.Env.GetIntOr("MAX_ROOMS", 1000000)

	hostname, err := os.Hostname()
	if err != nil {
		b.Log.Errorf("Couldn't get hostname, defaulting to localhost")
		hostname = "localhost"
	}

	WorkerQueue = make(chan chan WorkRequest, numWorkers)
	wg := &sync.WaitGroup{}

	internalWorkers := b.startInternalWorkers(numWorkers, wg, maxRoomsToProcess)

	// this is the server that picks up jobs from the queue
	taskServer := NewTaskServer()
	taskServer.RegisterTask("autoArchive", b.autoArchive)
	worker := taskServer.NewWorker(fmt.Sprintf("%s:machinery-worker", hostname))

	go b.handleExitSignal(internalWorkers, worker, wg)

	err = worker.Launch()
	if err != nil {
		panic(err)
	}
}

func (server *Server) startInternalWorkers(numWorkers int, wg *sync.WaitGroup, maxRoomsToProcess int) *[]Worker {
	internalWorkers := make([]Worker, numWorkers)

	for i := range internalWorkers {
		server.Log.Infof("Starting worker-%d", i+1)
		internalWorkers[i] = server.newWorker(i+1, WorkerQueue)
		internalWorkers[i].start(server, wg, maxRoomsToProcess)
	}

	return &internalWorkers
}

func (server *Server) handleExitSignal(internalWorkers *[]Worker, worker *machinery.Worker, wg *sync.WaitGroup) {
	sChan := make(chan os.Signal, 1)
	signal.Notify(sChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		// catch quit signal
		s := <-sChan
		server.Log.Infof("Signal %s received, stopping workers", s)

		for _, iw := range *internalWorkers {
			iw.stop()
		}

		wg.Wait()
		worker.Quit()

		return
	}()
}

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func (server *Server) newWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		Log:         server.Log,
	}

	return worker
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w Worker) start(s *Server, wg *sync.WaitGroup, maxRoomsToProcess int) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {

			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
				jobID := uuid.NewV4().String()
				w.Log.Infof("worker%d: Received work request for tid-%s", w.ID, work.TenantID)

				tenants := s.NewTenants()
				tenant, err := tenants.Get(work.TenantID)
				if err != nil {
					w.Log.Errorf("Coudn't find tid-%s", work.TenantID)
					continue
				}

				tenantConfigurations := s.NewTenantConfigurations()
				tenantConfiguration, err := tenantConfigurations.Get(tenant.ID)

				if err != nil {
					w.Log.Errorf("Coudn't find a configuration for tid-%s", work.TenantID)
					continue
				}

				credentials := hipchat.ClientCredentials{
					ClientID:     tenant.ID,
					ClientSecret: tenant.Secret,
				}

				newClient := hipchat.NewClient("")
				baseURL, _ := url.Parse(tenant.Links.API + "/")
				newClient.BaseURL = baseURL
				w.Log.Infof("NewClient.BaseURL %s", newClient.BaseURL)
				token, _, err := newClient.GenerateToken(
					credentials,
					[]string{hipchat.ScopeManageRooms, hipchat.ScopeViewGroup, hipchat.ScopeSendNotification, hipchat.ScopeAdminRoom})

				if err != nil {
					// this typically means the group uninstalled the plugin
					w.Log.Errorf("Client.GetAccessToken returns an error %v", err)
					continue
				}

				client := token.CreateClient()
				client.BaseURL = baseURL

				job := Job{
					Log:      w.Log.Record("jid", jobID).Record("tid", work.TenantID).Child(),
					JobID:    jobID,
					TenantID: work.TenantID,
					Client:   client,
				}

				rooms, error := job.GetRooms()
				if error != nil {
					w.Log.Errorf("Failed to retrieve rooms")
					continue
				}

				w.Log.Infof("Retrieved %d rooms", len(rooms))
				processedRooms := 0
				archivedRooms := 0
				for _, room := range rooms {
					archived := job.MaybeArchiveRoom(room.ID, tenantConfiguration.Threshold)
					if archived {
						archivedRooms++
					}
					processedRooms++
					if processedRooms%100 == 0 {
						w.Log.Infof("%d/%d rooms processed", processedRooms, len(rooms))
						w.Log.Infof("%d rooms archived so far", archivedRooms)
					}

					if processedRooms > maxRoomsToProcess {
						w.Log.Infof("Quota of %d rooms reached", maxRoomsToProcess)
						break
					}
				}

				w.Log.Infof("worker%d: Finished work request, archived %d rooms", w.ID, archivedRooms)

			case <-w.QuitChan:
				// We have been asked to stop.
				w.Log.Infof("worker-%d stopping", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (w Worker) stop() {
	go func() {
		w.QuitChan <- true
	}()
}

// Machinery requires to return a (interface{}, error) even if we don't handle
// the result, so faking it for now (shrug)
func (server *Server) autoArchive(tenantID string) (bool, error) {
	// this sends the autoArchive requests to one of the internal workers to process
	work := WorkRequest{tenantID}
	worker := <-WorkerQueue
	worker <- work

	return true, nil
}
