package main

var WorkerQueue chan chan WorkRequest

func StartDispatcher(context *Context, nworkers int) {
  // First, initialize the channel we are going to but the workers' work channels into.
  WorkerQueue = make(chan chan WorkRequest, nworkers)

  // Now, create all of our workers.
  for i := 0; i<nworkers; i++ {
    log.Debug("Starting worker", i+1)
    worker := NewWorker(i+1, WorkerQueue)
    worker.Start(context)
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
