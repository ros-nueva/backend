package main

import (
	"reflect"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"appengine"
	"appengine/datastore"
	"errors"
	"time"
)

var (
	ErrTripAlreadyExists = errors.New("trip already exists")
	ErrTripMissing = errors.New("trip does not exist")
)

type TripManager struct {
	manager *PubnubManager
}

type Trip struct {
	ID string `datastore:"id" json:"id"`
	JourneyID string `datastore:"journey_id" json:"journey_id"`
	Description string `datastore:"description" json:"description"`
	StartRoom string `datastore:"start_room" json:"start`
	EndRoom string `datastore:"end_room" json:"end"`
	Success bool `datastore:"success" json:"success"`
	LeftAt int64 `datastore:"left_at"`
	ArrivedAt int64 `datastore:"arrived_at"`
}

func (manager TripManager) Group() Group {
	return Group{
		Paths: Routes{
			"{tripid}/": Group{
				Paths: Routes{
					"get": Route{
						Handler: manager.GetTrip,
					},
					"create": Route{
						Handler: manager.CreateTrip,
					},
					"delete": Route{
						Handler: manager.DelTrip,
					},
					"set": Route{
						Handler: manager.SetTrip,
					},
					"start": Route{
						Handler: manager.StartTrip,
					},
					"complete": Route{
						Handler: manager.CompleteTrip,
					},
				},
			},
		},
	}
}

func (manager TripManager) CreateTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	trip := Trip{}
	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Get(ctx, key, &trip)
	if dataErr == nil {
		encoder.Encode(ResponseError{Error: ErrTripAlreadyExists.Error()})
		return
	} else if dataErr != datastore.ErrNoSuchEntity {
		encoder.Encode(ResponseError{Error: dataErr.Error()})
		return
	}
	decoder.Decode(&trip)
	journey := &Journey{}
	journeyKey := datastore.NewKey(ctx, "journey", trip.JourneyID, 0, nil)
	journeyDataErr := datastore.Get(ctx, journeyKey, journey)
	if journeyDataErr != nil {
		if journeyDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: journeyDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	trip.ID = tripID
	journey.Trips = append(journey.Trips, trip.ID)
	datastore.Put(ctx, key, trip)
	datastore.Put(ctx, journeyKey, journey)
	encoder.Encode(trip)
}

func (manager TripManager) GetTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)

	trip := Trip{}
	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Get(ctx, key, &trip)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	encoder.Encode(trip)
}

func (manager TripManager) DelTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)

	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Delete(ctx, key)
	if dataErr != nil {
		if dataErr != datastore.ErrInvalidKey {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	encoder.Encode(ResponseSuccess{Success: true})
}

func (manager TripManager) SetTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	trip := &Trip{}
	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Get(ctx, key, trip)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	updatedTrip := Trip{}
	decoder.Decode(updatedTrip)
	trip.MergeInPlace(updatedTrip)
	datastore.Put(ctx, key, trip)
	encoder.Encode(*trip)
}

func (manager TripManager) StartTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)

	trip := &Trip{}
	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Get(ctx, key, trip)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	journey := &Journey{}
	journeyKey := datastore.NewKey(ctx, "journey", trip.JourneyID, 0, nil)
	journeyDataErr := datastore.Get(ctx, journeyKey, journey)
	if journeyDataErr != nil {
		if journeyDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: journeyDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	journey.LatestTrip = trip.ID
	trip.LeftAt = time.Now().UTC().Unix()
	datastore.Put(ctx, key, *trip)
	datastore.Put(ctx, journeyKey, *journey)
	encoder.Encode(*trip)
}

func (manager TripManager) CompleteTrip(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tripID := mux.Vars(r)["tripid"]
	encoder := json.NewEncoder(w)

	trip := &Trip{}
	key := datastore.NewKey(ctx, "trip", tripID, 0, nil)
	dataErr := datastore.Get(ctx, key, trip)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	journey := &Journey{}
	journeyKey := datastore.NewKey(ctx, "journey", trip.JourneyID, 0, nil)
	journeyDataErr := datastore.Get(ctx, journeyKey, journey)
	if journeyDataErr != nil {
		if journeyDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: journeyDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	if trip.ID == journey.Trips[len(journey.Trips) - 1] {
		journey.Finished = true
	}
	trip.ArrivedAt = time.Now().UTC().Unix()
	trip.Success = true
	datastore.Put(ctx, key, *trip)
	encoder.Encode(*trip)
}

func (old *Trip) MergeInPlace(new Trip) {
	for ii := 0; ii < reflect.TypeOf(old).Elem().NumField(); ii++ {
		if x := reflect.ValueOf(&new).Elem().Field(ii); !x.IsNil() {
			reflect.ValueOf(old).Elem().Field(ii).Set(x)
		}
	}
}