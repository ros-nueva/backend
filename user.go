package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserMissing       = errors.New("user does not exist")
)

type UserManager struct{}

type User struct {
	ID            string   `datastore:"id" json:"id"`
	FirstName     string   `datastore:"first_name" json:"first_name"`
	LastName      string   `datastore:"last_name" json:"last_name"`
	Description   string   `datastore:"description" json:"description"`
	Likes         []string `datastore:"likes" json:"likes"`
	Grade         int      `datastore:"grade" json:"grade"`
	Journeys      []string `datastore:"journeys" json:"journeys"`
	LatestJourney string   `datastore:"latest_journey" json:"latest_journey"`
}

func (manager UserManager) Group() Group {
	return Group{
		Paths: Routes{
			"{userid}/": Group{
				Paths: Routes{
					"get": Route{
						Handler: manager.GetUser,
					},
					"create": Route{
						Handler: manager.CreateUser,
					},
					"delete": Route{
						Handler: manager.DelUser,
					},
					"set": Route{
						Handler: manager.SetUser,
					},
				},
			},
		},
	}
}

func (manager UserManager) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	userID := mux.Vars(r)["userid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	user := User{}
	key := datastore.NewKey(ctx, "user", userID, 0, nil)
	dataErr := datastore.Get(ctx, key, &user)
	if dataErr == nil {
		encoder.Encode(ResponseError{Error: ErrUserAlreadyExists.Error()})
		return
	} else if dataErr != datastore.ErrNoSuchEntity {
		encoder.Encode(ResponseError{Error: dataErr.Error()})
		return
	}
	decoder.Decode(&user)
	user.ID = userID
	datastore.Put(ctx, key, user)
	encoder.Encode(user)
}

func (manager UserManager) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	userID := mux.Vars(r)["userid"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	user := User{}
	key := datastore.NewKey(ctx, "user", userID, 0, nil)
	dataErr := datastore.Get(ctx, key, &user)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrUserMissing.Error()})
		return
	}
	decoder.Decode(&user)
	encoder.Encode(user)
}

func (manager UserManager) DelUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	userID := mux.Vars(r)["userid"]
	encoder := json.NewEncoder(w)

	key := datastore.NewKey(ctx, "user", userID, 0, nil)
	dataErr := datastore.Delete(ctx, key)
	if dataErr != nil {
		if dataErr != datastore.ErrInvalidKey {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrUserMissing.Error()})
	}
	encoder.Encode(ResponseSuccess{Success: true})
}

func (manager UserManager) SetUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	userID := mux.Vars(r)["userID"]
	encoder := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)

	user := User{}
	key := datastore.NewKey(ctx, "user", userID, 0, nil)
	dataErr := datastore.Get(ctx, key, &user)
	if dataErr != nil {
		if dataErr != datastore.ErrNoSuchEntity {
			encoder.Encode(ResponseError{Error: dataErr.Error()})
			return
		}
		encoder.Encode(ResponseError{Error: ErrUserMissing.Error()})
		return
	}
	updatedUser := User{}
	decoder.Decode(&updatedUser)
	user.MergeInPlace(updatedUser)
	datastore.Put(ctx, key, user)
	encoder.Encode(user)
}

func (old *User) MergeInPlace(new User) {
	for ii := 0; ii < reflect.TypeOf(old).Elem().NumField(); ii++ {
		if x := reflect.ValueOf(&new).Elem().Field(ii); !x.IsNil() {
			reflect.ValueOf(old).Elem().Field(ii).Set(x)
		}
	}
}
