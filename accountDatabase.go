package main

import (
	"encoding/json"
	"github.com/cockroachdb/pebble"
)

var accountDb *pebble.DB

func OpenAccountDb() {

	var err error
	accountDb, err = pebble.Open("users", &pebble.Options{})

	if err != nil {
		panic(err)
	}
}

func GetUser(username string) (User, error) {

	var user User
	value, _, err := accountDb.Get([]byte(username))

	if err != nil {
		return user, err
	}
	err = json.Unmarshal(value, &user)

	if err != nil {
		return user, err
	}
	return user, nil
}

func AddUser(username string, password string, imageId string) error {

	user := User{
		UserName: username,
		Password: password,
		Id:       imageId,
	}

	userBytes, err := json.Marshal(user)

	if err != nil {
		return err
	}
	err = accountDb.Set([]byte(username), userBytes, pebble.Sync)

	if err != nil {
		return err
	}

	return nil
}
