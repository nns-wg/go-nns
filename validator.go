package main

import record "github.com/libp2p/go-libp2p-record"

var _ record.Validator = Validator{}

func (v Validator) Validate(key string, value []byte) error {

}

func (v Validator) Select(k, string, vals [][]byte) (int, error) {
	// fixme this is obviously incorrect and will always return the first record, not the most recent / etc
	return 0
}
