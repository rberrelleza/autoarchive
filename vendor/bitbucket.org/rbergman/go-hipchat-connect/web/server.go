package web

import (
	"errors"
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
	"bitbucket.org/rbergman/go-hipchat-connect/util"

	"github.com/chakrit/go-bunyan"
	"github.com/codegangsta/negroni"
	"github.com/garyburd/redigo/redis"
	"github.com/go-zoo/bone"
)

// Server represents an HTTP server instance with Negroni middleware and Bone
// routing.
type Server struct {
	AppName    string
	Port       string
	Log        bunyan.Log
	Router     *bone.Mux
	Middleware *negroni.Negroni
	Descriptor string
	RedisPool  *redis.Pool
}

// NewServer creates a new Server with the given app name.
func NewServer(descriptorPath, publicPath string) *Server {
	port := util.Env.GetStringOr("PORT", "3000")
	baseURL := util.Env.GetStringOr("BASE_URL", "http://localhost:"+port)
	desc := model.MustLoadDescriptor(descriptorPath, baseURL)
	if desc.Key == "" {
		panic("Required descriptor field 'key' is undefined.")
	}
	name := desc.Key
	log := NewStdLogger(name)
	router := bone.New()
	mw := NewMiddleware(name, publicPath)

	return &Server{
		AppName:    name,
		Port:       port,
		Log:        log,
		Router:     router,
		Middleware: mw,
		Descriptor: model.MustString(desc),
		RedisPool:  newRedisPool(),
	}
}

// Start starts the current server by listening on all addresses at the port
// specificed by the PORT env var or 3000, othwerwise.
func (s *Server) Start() {
	s.Middleware.UseHandler(s.Router)
	address := ":" + s.Port
	s.Log.Record("address", address).Infof("listening")
	http.ListenAndServe(address, s.Middleware)
}

// TODO turn this into route-handler-specific middleware
func (s *Server) VerifyJSONRequest(req *http.Request) (int, error) {
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		return http.StatusUnsupportedMediaType, errors.New("Unsupported media type: " + contentType)
	}
	if req.Body == nil {
		return http.StatusBadRequest, errors.New("Expected JSON body")
	}
	return 0, nil
}
