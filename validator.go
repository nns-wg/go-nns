package main

import (
	"context"
	"log"

	"github.com/nns-wg/go-nns/whocan"
	"github.com/ucan-wg/go-ucan"
)

type Validator struct {
	ctx context.Context
}

func (v Validator) Validate(key string, value []byte) error {

	store := ucan.NewMemTokenStore()
	// FIXME this is broken, but I'm not quite sure how to fix it. This may be an
	// issue of using an old UCAN spec in the go lib. Really, we need to figure out what the capability is!
	ac := func(m map[string]interface{}) (ucan.Attenuation, error) {
		var rsc ucan.Resource
		//		var cap string

		log.Printf("attempting to build attenuations out of %v", m)
		// is this supposed to be "with"
		rsc = ucan.NewStringLengthResource("ucan", "*")
		// and this "can"?
		// the formats don't seem to match
		caps := ucan.NewNestedCapabilities("DELEGATE")
		return ucan.Attenuation{Rsc: rsc, Cap: caps.Cap("DELEGATE")}, nil
	}

	parser := whocan.NewTokenParser(ac, whocan.GenericDIDPubKeyResolver{}, store.(ucan.CIDBytesResolver))
	tok, err := parser.ParseAndVerify(v.ctx, string(value[:]))
	if err != nil {
		log.Printf("Unable to parse/verify token: %+v", err)
		return err
	}

	log.Printf(" the token: %+v", tok)
	return nil
}

func (v Validator) Select(k string, vals [][]byte) (int, error) {
	// fixme this is obviously incorrect and will always return the first record, not the most recent / etc
	log.Printf("In select %v: %v", k, vals)
	return 0, nil
}
