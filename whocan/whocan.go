/*

This is all lifted from go-ucan, with the intent of merging it back in.

The primary difference is the DIDPubKeyResolver interface needs access to the
token in order to resolve DIDs that contain supporting information in the fct
field.

*/

package whocan

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	resolvers "github.com/nns-wg/go-nns/resolvers"
	"github.com/ucan-wg/go-ucan"
	"github.com/ucan-wg/go-ucan/didkey"
)

type DIDPubKeyResolver interface {
	ResolveDIDKey(ctx context.Context, did string, tok *jwt.Token) (didkey.ID, error)
}

type Token struct {
	// Entire UCAN as a signed JWT string
	Raw string
	// the resolved did:keys corresponding to the raw DIDs
	Issuer   didkey.ID
	Audience didkey.ID
	// the raw DIDs (e.g., did:mailto:user@example.com)
	IssuerDID   string
	AudienceDID string
	// the "inputs" to this token, a chain UCAN tokens with broader scopes &
	// deadlines than this token
	Proofs []ucan.Proof `json:"prf,omitempty"`
	// the "outputs" of this token, an array of heterogenous resources &
	// capabilities
	Attenuations ucan.Attenuations `json:"att,omitempty"`
	// Facts are facts, jack.
	Facts []ucan.Fact `json:"fct,omitempty"`
}

// TokenParser parses a raw string into a Token
type TokenParser struct {
	ap   ucan.AttenuationConstructorFunc
	cidr ucan.CIDBytesResolver
	didr DIDPubKeyResolver
}

// NewTokenParser constructs a token parser
func NewTokenParser(ap ucan.AttenuationConstructorFunc, didr DIDPubKeyResolver, cidr ucan.CIDBytesResolver) *TokenParser {
	return &TokenParser{
		ap:   ap,
		cidr: cidr,
		didr: didr,
	}
}

func (p *TokenParser) ParseAndVerify(ctx context.Context, raw string) (*Token, error) {
	return p.parseAndVerify(ctx, raw, nil)
}

func (p *TokenParser) parseAndVerify(ctx context.Context, raw string, child *Token) (*Token, error) {

	tok, err := jwt.Parse(raw, p.matchVerifyFunc(ctx))
	if err != nil {
		return nil, fmt.Errorf("parsing UCAN: %w", err)
	}

	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("parser fail")
	}

	var issKey didkey.ID
	var issDID string

	if issStr, ok := mc["iss"].(string); ok {
		issKey, err = p.didr.ResolveDIDKey(ctx, issStr, tok)
		//resolveDidKeyFor(issStr, tok)
		if err != nil {
			return nil, err
		}
		issDID = issStr
	} else {
		return nil, fmt.Errorf(`"iss" key is not in claims`)
	}

	var audKey didkey.ID
	var audDID string
	if audStr, ok := mc["aud"].(string); ok {
		audKey, err = p.didr.ResolveDIDKey(ctx, audStr, tok)
		if err != nil {
			return nil, err
		}
		audDID = audStr
	} else {
		return nil, fmt.Errorf(`"aud" key is not in claims`)
	}

	var att ucan.Attenuations
	if acci, ok := mc[ucan.AttKey].([]interface{}); ok {
		for i, a := range acci {
			if mapv, ok := a.(map[string]interface{}); ok {
				a, err := p.ap(mapv)
				if err != nil {
					return nil, err
				}
				att = append(att, a)
			} else {
				return nil, fmt.Errorf(`"att[%d]" is not an object`, i)
			}
		}
	} else {
		return nil, fmt.Errorf(`"att" key is not an array`)
	}

	var prf []ucan.Proof
	if prfi, ok := mc[ucan.PrfKey].([]interface{}); ok {
		for i, a := range prfi {
			if pStr, ok := a.(string); ok {
				prf = append(prf, ucan.Proof(pStr))
			} else {
				return nil, fmt.Errorf(`"prf[%d]" is not a string`, i)
			}
		}
	} else if mc[ucan.PrfKey] != nil {
		return nil, fmt.Errorf(`"prf" key is not an array`)
	}

	return &Token{
		Raw:          raw,
		Issuer:       issKey,
		IssuerDID:    issDID,
		Audience:     audKey,
		AudienceDID:  audDID,
		Attenuations: att,
		Proofs:       prf,
	}, nil
}

func (p *TokenParser) matchVerifyFunc(ctx context.Context) func(tok *jwt.Token) (interface{}, error) {

	return func(tok *jwt.Token) (interface{}, error) {
		mc, ok := tok.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("parser fail")
		}

		iss, ok := mc["iss"].(string)
		if !ok {
			return nil, fmt.Errorf(`"iss" claims key is required`)
		}

		id, err := p.didr.ResolveDIDKey(ctx, iss, tok)
		if err != nil {
			return nil, err
		}

		return id.VerifyKey()
	}
}

type GenericDIDPubKeyResolver struct{}

func (GenericDIDPubKeyResolver) ResolveDIDKey(ctx context.Context, iss string, tok *jwt.Token) (didkey.ID, error) {

	switch {
	case strings.HasPrefix(iss, "did:key:"):
		return didkey.Parse(iss)

	// OKAY, SO.
	//
	// These aren't standardized DIDs. That should probably change. BUT.
	// BUT.
	//
	// The DID universal resolver requires Docker. Or shelling out to a
	// centralized HTTPS endpoint (for testing purposes only).
	//
	// Adding universal DID support "should" be easy. But I don't want to
	// spend my time right now building out DID support (docker, etc), but
	// rather just prove out some really simple use-cases, and not worry about
	// conforming to (in many cases, half-baked) specs.
	//
	// ALSO there's a question of DID -> Identifier
	// normalization/canonicalization, and the advantage of using opaque URIs
	// here is that we don't have to figure that part out right now.
	//
	// ALSO for the purposes of NNS, there are multiple ways to verify ownership
	// of a key. The Keybase mechanisms are fine, too, in addition to DID methods.
	//
	// So I'm doing these custom schemes, with the expectation that where
	// appropriate, formal DID methods can be supported.
	case strings.HasPrefix(iss, "did:mailto:"):
		return resolvers.ResolveMailtoKey(iss, tok)
		//case strings.HasPrefix(iss, "https:"):
		//  return p.resolveHttpKey(iss, tok)
		//case strings.HasPrefix(iss, "dnssec:"):
		//	return p.resolveDNSKey(iss, tok)
	}

	var id didkey.ID
	return id, fmt.Errorf("no supported verification scheme for %s", iss)
}

/*
func resolveHttpKey(iss string, tok *jwt.Token) (didkey.ID, error) {

}
*/
