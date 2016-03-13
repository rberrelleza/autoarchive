package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/web"
)

type Server struct {
	web.Server
}

func NewServer() *Server {
	s := &Server{*web.NewServer("./static/descriptor.json")}
	s.AddGetConfigurable(s.configurable)
	s.AddPostConfigurable(s.postConfigurable)
	return s
}

func (c *Context) RunWeb() {
	NewServer().Start()
}
