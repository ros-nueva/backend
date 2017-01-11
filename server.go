package main

import (
	"net/http"
	"log"
)

type Server struct {
	UserManager
	JourneyManager
	TripManager
	RoomManager
	*PubnubManager
}

type Manager interface {
	Group() Group
}

type ResponseError struct {
	Error string `json:"error"`
}

type ResponseSuccess struct {
	Success bool `json:"success"`
}

func (server Server) Handler() http.Handler {
	router := Routes{
		"trip/": server.TripManager.Group(),
		"user/": server.UserManager.Group(),
		"room/": server.RoomManager.Group(),
		"journey/": server.JourneyManager.Group(),
	}
	return router.Serve()
}

func (server Server) MiddleEncoding(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Println("hit")
		log.Println(r)
		f(w, r)
	}
}