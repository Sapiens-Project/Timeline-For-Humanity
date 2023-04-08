package main

import (
	"github.com/akrylysov/pogreb"
)

type Cache string

func (c Cache) Put(key, val []byte) error {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Put([]byte(key), val)
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
