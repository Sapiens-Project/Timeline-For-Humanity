package main

import (
	"errors"

	"github.com/akrylysov/pogreb"
)

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

// Fold iterates over all the key-value pairs stored in the cache and calls the
// function in input passing those values as argument.
func (c Cache) Fold(fn func(key, val []byte) error) (err error) {
	cc, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer cc.Close()

	iter := cc.Items()
	for err == nil {
		if key, val, err := iter.Next(); err == nil {
			err = fn(key, val)
		}
	}

	if errors.Is(err, pogreb.ErrIterationDone) {
		err = nil
	}
	return
}

// Compact reduces the size of the cache on the disk.
func (c Cache) Compact() error {
	cc, err := pogreb.Open(string(c), nil)
	if err != nil {
		return err
	}
	defer cc.Close()
	_, err = cc.Compact()
	return err
}
