package main

import (
	"log"
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	pubnubManager := &PubnubManager{}
	server := Server{
		PubnubManager: pubnubManager,
		JourneyManager: JourneyManager{pubnubManager},
		TripManager: TripManager{pubnubManager},
	}
	server.Initialize()
	log.Println("Backend initializing...")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/user/{userid}/get", server.UserManager.GetUser)
	router.HandleFunc("/user/{userid}/create", server.UserManager.CreateUser)
	router.HandleFunc("/user/{userid}/del", server.UserManager.DelUser)
	router.HandleFunc("/user/{userid}/set", server.UserManager.SetUser)

	router.HandleFunc("/journey/{journeyid}/get", server.JourneyManager.GetJourney)
	router.HandleFunc("/journey/{journeyid}/create", server.JourneyManager.CreateJourney)
	router.HandleFunc("/journey/{journeyid}/del", server.JourneyManager.DelJourney)
	router.HandleFunc("/journey/{journeyid}/set", server.JourneyManager.SetJourney)
	router.HandleFunc("/journey/{journeyid}/set", server.JourneyManager.StartJourney)
	router.HandleFunc("/journey/{journeyid}/complete", server.JourneyManager.CompleteJourney)

	router.HandleFunc("/trip/{tripid}/get", server.TripManager.GetTrip)
	router.HandleFunc("/trip/{tripid}/create", server.TripManager.CreateTrip)
	router.HandleFunc("/trip/{tripid}/del", server.TripManager.DelTrip)
	router.HandleFunc("/trip/{tripid}/set", server.TripManager.SetTrip)
	router.HandleFunc("/trip/{tripid}/complete", server.TripManager.CompleteTrip)
	router.HandleFunc("/trip/{tripid}/start", server.TripManager.StartTrip)

	router.HandleFunc("/room/{roomid}/get", server.RoomManager.GetRoom)
	router.HandleFunc("/room/{roomid}/create", server.RoomManager.CreateRoom)
	router.HandleFunc("/room/{roomid}/del", server.RoomManager.DelRoom)
	router.HandleFunc("/room/{roomid}/set", server.RoomManager.SetRoom)

	http.Handle("/", router)

	log.Println("Backend initialized...")
}
