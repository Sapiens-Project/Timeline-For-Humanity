package main

import "github.com/akrylysov/pogreb"

type Cache string

func (c Cache) Put(key, val []byte) error {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Put(key, val)
	if err != nil {
		return err
	}
	return nil
}

func (c Cache) Get(key []byte) ([]byte, error) {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c Cache) Del(key []byte) error {
	db, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Delete(key)
}
