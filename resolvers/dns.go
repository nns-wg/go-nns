package resolvers

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/ucan-wg/go-ucan/didkey"
)

const hostnameRegexp = `^(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}|\.?)$`

func ResolveDnsKey(iss string, tok *jwt.Token) (didkey.ID, error) {
	var id didkey.ID

	issAddr := iss[len("did:dns:"):]

	didRecords, err := net.LookupTXT("_did." + issAddr)
	if err != nil {
		return id, fmt.Errorf("unable to find dns proof: %v %v", didRecords, err)
	}

	// FIXME it's not necessarily true that the first record is the one we want
	// we should check that the record matches the signature (but this isn't a
	// problem in general because the UCAN won't be valid if the signature is
	// invalid)
	for _, didRecord := range didRecords {
		log.Printf("did record! %v", didRecord)
		key, err := didkey.Parse(didRecord)

		if err != nil {
			log.Printf("unable to parse did record: %v", err)
			continue
		}
		
	  return key, nil
	}

	return id, fmt.Errorf("unable to find dns proof: %v", err)
}

func MatchDnsIssuer(iss string, name string) (bool) {

	if !strings.HasPrefix(iss, "did:dns:") {
		return false
	}

  issAddr := iss[len("did:odd:"):]

	re := regexp.MustCompile(hostnameRegexp)

	isValid := net.ParseIP(strings.TrimRight(name, ".")) == nil &&
	           len(name) <= 253 &&
						 strings.Count(name, ".") <= 126 &&
						 re.MatchString(name)

  if !isValid {
		return false
	}

	if name == issAddr || name == "dns:" + issAddr {
		return true
	}

	return false
}