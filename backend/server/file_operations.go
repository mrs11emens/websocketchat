package server

import (
	"encoding/json"
	"io/ioutil"
)

const (
	userFile    = "users.json"
	messageFile = "messages.json"
)

func loadUsers() (UserList, error) {
	file, err := ioutil.ReadFile(userFile)
	if err != nil {
		return UserList{}, err
	}

	var users UserList
	err = json.Unmarshal(file, &users)
	if err != nil {
		return UserList{}, err
	}

	return users, nil
}

func saveUsers(users UserList) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(userFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadMessages() (MessageHistory, error) {
	file, err := ioutil.ReadFile(messageFile)
	if err != nil {
		return MessageHistory{}, err
	}

	var history MessageHistory
	err = json.Unmarshal(file, &history)
	if err != nil {
		return MessageHistory{}, err
	}

	return history, nil
}

func saveMessages(history MessageHistory) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(messageFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}