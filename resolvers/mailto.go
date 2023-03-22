package resolvers

import (
	"bytes"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strings"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/golang-jwt/jwt"
	"github.com/ucan-wg/go-ucan/didkey"
)

const subjectRegexp = `^VERIFYING ([^@]+@[^\.]+\.[^ ]+) IS OWNED BY (did:key:[a-zA-Z0-9]+)$`

func ResolveMailtoKey(iss string, tok *jwt.Token) (didkey.ID, error) {

	var id didkey.ID

	rawMessage, err := extractDKIMProof(tok)
	if err != nil {
		return id, fmt.Errorf("unable to find dkim proof: %v", err)
	}

	key, err := verifyMailMessage(iss, rawMessage)
	if err != nil {
		return id, fmt.Errorf("invalid mail delegation: %v", err)
	}

	err = verifyMessageDKIM(rawMessage)
	if err != nil {
		return id, fmt.Errorf("invalid dkim proof: %v", err)
	}

	return didkey.Parse(key)
}

func extractDKIMProof(tok *jwt.Token) (string, error) {

	var dkimProof string
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return dkimProof, fmt.Errorf("parser fail")
	}

	fct, ok := mc["fct"]
	if !ok {
		return dkimProof, fmt.Errorf(`"fct" key is required in order to verify mailto aud`)
	}

	maybeDkimProof, ok := fct.([]interface{})[0].(map[string]interface{})["dkimProof"]
	if !ok {
		return dkimProof, fmt.Errorf("dkimProof must be present in fct in order to verify mailto issuer")
	}
	dkimProof = maybeDkimProof.(string)

	return dkimProof, nil

}

func verifyMailMessage(iss string, rawMessage string) (string, error) {

	var key string

	mailReader := bytes.NewBufferString(rawMessage)
	msg, err := mail.ReadMessage(mailReader)
	if err != nil {
		return key, fmt.Errorf("unable to parse email %v %v", err, rawMessage)
	}

	issAddr := iss[len("did:mailto:"):]
	fromHeader := msg.Header.Get("from")
	fromAddr, err := mail.ParseAddress(fromHeader)
	if err != nil {
		return key, fmt.Errorf("unable to parse from address %v %s", err, fromHeader)
	}

	if issAddr != fromAddr.Address {
		return key, fmt.Errorf("%s does not match from address %s", issAddr, fromAddr.Address)
	}

	subject := msg.Header.Get("subject")

	subjectRE := regexp.MustCompile(subjectRegexp)
	matches := subjectRE.FindStringSubmatch(subject)

	if len(matches) != 3 {
		return key, fmt.Errorf("no valid delegation found in email subject header (must match ^VERIFYING <address> IS OWNED BY <did:key:...>$). Got header '%s'", subject)
	}
	delegateAddr := matches[1]
	if issAddr != delegateAddr {
		return key, fmt.Errorf("-->%s<-- does not match Subject delegation address -->%s<--", iss, delegateAddr)
	}

	return matches[2], nil
}

func verifyMessageDKIM(rawMessage string) error {

	proof := bytes.NewBufferString(rawMessage)
	verifications, err := dkim.Verify(proof)
	if err != nil {
		return fmt.Errorf("unable to verify dkim proof %v", err)
	}

	for _, v := range verifications {
		if v.Err == nil {

			var subjectSigned bool = false

			for _, header := range v.HeaderKeys {
				log.Printf("header %v", header)
				if header == "subject" {
					subjectSigned = true
					break
				}
			}

			if !subjectSigned {
				return fmt.Errorf("subject not verified")
			}

		} else {
			return fmt.Errorf("invalid dkim signature")
		}
	}

	return nil
}

func MatchMailtoIssuer(iss string, name string) (bool) {
  if !strings.HasPrefix(iss, "did:mailto:") {
		return false
	}

  issAddr := iss[len("did:mailto:"):]

	log.Printf("name: %s, iss: %s, issAddr: %s", name, iss, issAddr)

	switch name {
		case issAddr: return true // this likely needs some validation

		// as do these, but more worried about the first one
		case "mailto:" + issAddr: return true
		case "acct:" + issAddr: return true
	}

	return false
}