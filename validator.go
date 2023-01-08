package main

import "log"


type Validator struct {}

func (v Validator) Validate(key string, value []byte) error {
	log.Printf("I'm in here? %s", key)
  return nil
}

func (v Validator) Select(k string, vals [][]byte) (int, error) {
	// fixme this is obviously incorrect and will always return the first record, not the most recent / etc
	log.Printf("In select %v: %v", k, vals)
	return 0, nil
}