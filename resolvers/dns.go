package resolvers

import (
	"github.com/golang-jwt/jwt"
	"github.com/ucan-wg/go-ucan/didkey"
)

func ResolveDnsKey(iss string, tok *jwt.Token) (didkey.ID, error) {
	var id didkey.ID
	return id, nil
}

func MatchDnsIssuer(iss string, name string) (bool) {
	return false
}