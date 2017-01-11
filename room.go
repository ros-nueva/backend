package main

import (
	"reflect"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"appengine"
	"appengine/datastore"
	"errors"
)

var (
	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrRoomMissing = errors.New("room does not exist")
)

type RoomManager struct {}

type Room struct {
	ID string `datastore:"id" json:"id"`
	Name string `datastore:"name" json:"name"`
	Description string `datastore:"description" json:"description"`
	Pose struct {
		Floor int `datastore:"z_pos" json:"z"`
		X int `datastore:"x_pos" json:"x"`
		Y int `datastore:"y_pos" json:"y"`
	} `datastore:"pose" json:"pose"`
}

func (manager RoomManager) Group() Group {
	return Group{
		Paths: Routes{
			"{roomid}/": Group{
				Paths: Routes{
					"get": Route{
						Handler: manager.GetRoom,
					},
					"create": Route{
						Handler: manager.CreateRoom,
					},
					"delete": Route{
						Handler: manager.DelRoom,
					},
					"set": Route{
						Handler: manager.SetRoom,
					},
				},
			},
		},
	}
}

func (manager RoomManager) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	roomID := mux.Vars(r)["roomid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	room := Room{}
	key := datastore.NewKey(ctx, "room", roomID, 0, nil)
	dataErr := datastore.Get(ctx, key, &room)
	if dataErr == nil {
		encoder.Encode(ResponseError{Error: ErrUserAlreadyExists.Error()})
		return
	} else if dataErr != datastore.ErrNoSuchEntity {
		encoder.Encode(ResponseError{Error: dataErr.Error()})
		return
	}
	decoder.Decode(&room)
	room.ID = roomID
	datastore.Put(ctx, key, room)
	encoder.Encode(room)
}

func (manager RoomManager) GetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	roomID := mux.Vars(r)["roomid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	room := Room{}
	key := datastore.NewKey(ctx, "room", roomID, 0, nil)
	dataErr := datastore.Get(ctx, key, &room)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrRoomMissing.Error()})
		return
	}
	decoder.Decode(&room)
	encoder.Encode(room)
}

func (manager RoomManager) DelRoom(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	roomID := mux.Vars(r)["roomid"]
	encoder := json.NewEncoder(w)

	key := datastore.NewKey(ctx, "room", roomID, 0, nil)
	dataErr := datastore.Delete(ctx, key)
	if dataErr != nil {
		if dataErr != datastore.ErrInvalidKey {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrRoomMissing.Error()})
	}
	encoder.Encode(ResponseSuccess{Success: true})
}

func (manager RoomManager) SetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	roomID := mux.Vars(r)["roomid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	room := Room{}
	key := datastore.NewKey(ctx, "room", roomID, 0, nil)
	dataErr := datastore.Get(ctx, key, &room)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrRoomMissing.Error()})
		return
	}
	updatedRoom := Room{}
	decoder.Decode(&updatedRoom)
	room.MergeInPlace(updatedRoom)
	datastore.Put(ctx, key, room)
	encoder.Encode(room)
}

func (old *Room) MergeInPlace(new Room) {
	for ii := 0; ii < reflect.TypeOf(old).Elem().NumField(); ii++ {
		if x := reflect.ValueOf(&new).Elem().Field(ii); !x.IsNil() {
			reflect.ValueOf(old).Elem().Field(ii).Set(x)
		}
	}
}