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
	ErrJourneyAlreadyExists = errors.New("journey already exists")
	ErrJourneyMissing = errors.New("journey does not exist")
)

type JourneyManager struct {
	pubnubManager *PubnubManager
}

type Journey struct {
	ID string `datastore:"id" json:"id"`
	User string `datastore:"user_id" json:"user_id"`
	Name string `datastore:"name" json:"name"`
	StartAt int64 `datastore:"start_time" json:"start_at"`
	FinishedAt int64 `datastore:"finished_time" json:"finished_at"`
	Trips []string `datastore:"trips" json:"trips"`
	LatestTrip string `datastore:"latest_trip" json:"latest_trip"`
	Finished bool `datastore:"finished"`
}

func (manager JourneyManager) Group() Group {
	return Group{
		Paths: Routes{
			"{journeyid}/": Group{
				Paths: Routes{
					"get": Route{
						Handler: manager.GetJourney,
					},
					"create": Route{
						Handler: manager.CreateJourney,
					},
					"delete": Route{
						Handler: manager.DelJourney,
					},
					"set": Route{
						Handler: manager.SetJourney,
					},
					"complete": Route{
						Handler: manager.CompleteJourney,
					},
				},
			},
		},
	}
}

func (manager JourneyManager) CreateJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	journey := Journey{}
	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Get(ctx, key, &journey)
	if dataErr == nil {
		encoder.Encode(ResponseError{Error: ErrJourneyAlreadyExists.Error()})
		return
	} else if dataErr != datastore.ErrNoSuchEntity {
		encoder.Encode(ResponseError{Error: dataErr.Error()})
		return
	}
	decoder.Decode(&journey)
	user := &User{}
	userKey := datastore.NewKey(ctx, "user", journey.User, 0, nil)
	userDataErr := datastore.Get(ctx, userKey, user)
	if userDataErr != nil {
		if userDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: userDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrUserMissing.Error()})
		return
	}
	user.Journeys = append(user.Journeys, journey.ID)
	journey.ID = journeyID
	datastore.Put(ctx, key, journey)
	datastore.Put(ctx, userKey, user)
	encoder.Encode(journey)
}

func (manager JourneyManager) GetJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	journey := Journey{}
	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Get(ctx, key, &journey)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	decoder.Decode(&journey)
	encoder.Encode(journey)
}

func (manager JourneyManager) DelJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)

	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Delete(ctx, key)
	if dataErr != nil {
		if dataErr != datastore.ErrInvalidKey {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	encoder.Encode(ResponseSuccess{Success: true})
}

func (manager JourneyManager) SetJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	journey := &Journey{}
	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Get(ctx, key, journey)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	updatedJourney := Journey{}
	decoder.Decode(updatedJourney)
	journey.MergeInPlace(updatedJourney)
	datastore.Put(ctx, key, journey)
	encoder.Encode(*journey)
}

func (manager JourneyManager) StartJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)

	journey := &Journey{}
	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Get(ctx, key, journey)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	user := &User{}
	userKey := datastore.NewKey(ctx, "user", journey.User, 0, nil)
	userDataErr := datastore.Get(ctx, userKey, user)
	if userDataErr != nil {
		if userDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: userDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrUserMissing.Error()})
		return
	}
	trip := &Trip{}
	tripKey := datastore.NewKey(ctx, "trip", journey.LatestTrip, 0, nil)
	tripDataErr := datastore.Get(ctx, tripKey, trip)
	if tripDataErr != nil {
		if tripDataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: tripDataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrTripMissing.Error()})
		return
	}
	user.LatestJourney = journey.ID
	trip.LeftAt = time.Now().UTC().Unix()
	journey.StartAt = time.Now().UTC().Unix()
	manager.pubnubManager.PublishJSON(MessageStart{UserID: user.ID, JourneyID: journey.ID})
	datastore.Put(ctx, tripKey, tripKey)
	datastore.Put(ctx, key, journey)
	datastore.Put(ctx, key, journey)
	encoder.Encode(*journey)
}

func (manager JourneyManager) CompleteJourney(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	journeyID := mux.Vars(r)["journeyid"]
	encoder := json.NewEncoder(w)

	journey := &Journey{}
	key := datastore.NewKey(ctx, "journey", journeyID, 0, nil)
	dataErr := datastore.Get(ctx, key, journey)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrJourneyMissing.Error()})
		return
	}
	journey.Finished = true
	datastore.Put(ctx, key, journey)
	encoder.Encode(*journey)
}

func (old *Journey) MergeInPlace(new Journey) {
	for ii := 0; ii < reflect.TypeOf(old).Elem().NumField(); ii++ {
		if x := reflect.ValueOf(&new).Elem().Field(ii); !x.IsNil() {
			reflect.ValueOf(old).Elem().Field(ii).Set(x)
		}
	}
}