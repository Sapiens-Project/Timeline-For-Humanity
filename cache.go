package main

import (
	"encoding/json"

	"github.com/akrylysov/pogreb"
)

type Cache string

func (c Cache) Put(key, val string) error {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Put([]byte(key), []byte(val))
	if err != nil {
		return err
	}
	return nil
}

func (c Cache) Get(key string) ([]byte, error) {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	val, err := db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c Cache) Del(tlKey, dotKey string) error {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if dotKey == "" {
		err = db.Delete([]byte(tlKey))
		if err != nil {
			return err
		}
		return nil
	}

	if err = deleteDot(tlKey, dotKey, c); err != nil {
		return err
	}
	return nil
}

func deleteDot(tlKey, dotKey string, c Cache) error {
	var tl Timeline

	jsn, err := c.Get(tlKey)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsn, &tl); err != nil {
		return err
	}
	delete(tl.Dots, dotKey)

	newJsn, err := json.Marshal(tl)
	if err != nil {
		return err
	}
	c.Put(tlKey, string(newJsn))
	return nil
}
