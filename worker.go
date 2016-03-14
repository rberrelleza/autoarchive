package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"fmt"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var WorkerQueue chan chan WorkRequest

func StartWorker() {
	b := NewBackendServer("hiparchiver.workers")
	numWorkers := util.Env.GetIntOr("WORKERS_ENV", 6)

	hostname, err := os.Hostname()
	if err != nil {
		b.Log.Errorf("Couldn't get hostname, defaulting to localhost")
		hostname = "localhost"
	}

	WorkerQueue = make(chan chan WorkRequest, numWorkers)
	wg := &sync.WaitGroup{}

	internalWorkers := b.startInternalWorkers(numWorkers, wg)

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

func (server *Server) startInternalWorkers(numWorkers int, wg *sync.WaitGroup) *[]Worker {
	internalWorkers := make([]Worker, numWorkers)

	for i, _ := range internalWorkers {
		server.Log.Infof("Starting worker-%d", i+1)
		internalWorkers[i] = server.newWorker(i+1, WorkerQueue)
		internalWorkers[i].start(server, wg)
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
func (w Worker) start(s *Server, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {

			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.
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
				token, _, err := newClient.GenerateToken(
					credentials,
					[]string{hipchat.ScopeManageRooms, hipchat.ScopeViewGroup, hipchat.ScopeSendNotification, hipchat.ScopeAdminRoom})

				if err != nil {
					// this typically means the group uninstalled the plugin
					w.Log.Errorf("Client.GetAccessToken returns an error %v", err)
					continue
				}

				client := token.CreateClient()
				rooms, error := w.GetRooms(client)
				if error != nil {
					w.Log.Errorf("Failed to retrieve rooms for tid-%s", work.TenantID)
					continue
				}

				for _, room := range rooms {
					w.MaybeArchiveRoom(work.TenantID, room.ID, tenantConfiguration.Threshold, client)
				}

				w.Log.Infof("worker%d: Finished work request for tid-%s", w.ID, work.TenantID)

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
func (s *Server) autoArchive(tenantID string) (bool, error) {
	// this sends the autoArchive requests to one of the internal workers to process
	work := WorkRequest{tenantID}
	worker := <-WorkerQueue
	worker <- work

	return true, nil
}
