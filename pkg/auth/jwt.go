package auth

import (
	"context"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	signingKey                      = []byte(os.Getenv("JWT_TOKEN"))
	signingMethod jwt.SigningMethod = jwt.SigningMethodHS256
	issuer                          = "gideon"
	audience                        = "anyone"
	defaultAPI                      = &authAPI{signingKey: signingKey, audience: audience, issuer: issuer}
)

// Payload contains jwt payload
type Payload struct {
	ID           string
	ProjectID    string
	Names        string
	PhoneNumber  string
	EmailAddress string
	Group        string
}

// Claims contains JWT claims information
type Claims struct {
	*Payload
	jwt.StandardClaims
}

// AuthenticateRequest authenticates a request whether it contains valid jwt in metadata
func AuthenticateRequest(ctx context.Context) error {
	return defaultAPI.AuthenticateRequest(ctx)
}

// AuthenticateActor authenticates actor
func AuthenticateActor(ctx context.Context, actorID string) (*Payload, error) {
	return defaultAPI.AuthorizeActor(ctx, actorID)
}

// AuthorizeGroup authorizes an actor group against allowed groups
func AuthorizeGroup(ctx context.Context, allowedGroups ...string) (*Payload, error) {
	return defaultAPI.AuthorizeGroups(ctx, allowedGroups...)
}

// AuthorizeStrict authenticates and authorizes an actor and group against allowed groups
func AuthorizeStrict(ctx context.Context, actorID string, allowedGroups ...string) (*Payload, error) {
	return defaultAPI.AuthorizeStrict(ctx, actorID, allowedGroups...)
}

// AuthorizeActorOrGroup authorizes the actor or whether they belong to list of allowed groups
func AuthorizeActorOrGroup(ctx context.Context, actorID string, allowedGroups ...string) (*Payload, error) {
	return defaultAPI.AuthorizeActorOrGroups(ctx, actorID, allowedGroups...)
}

// GenToken generates jwt
func GenToken(ctx context.Context, payload *Payload, expirationTime time.Time) (string, error) {
	return defaultAPI.GenToken(ctx, payload, expirationTime)
}

// AddMD adds metadata to token
func AddMD(ctx context.Context, actorID, group string) context.Context {
	return defaultAPI.AddMD(ctx, actorID, group)
}

// ParseToken parses a jwt token and return claims
func ParseToken(tokenString string) (claims *Claims, err error) {
	return defaultAPI.ParseToken(tokenString)
}

// ParseFromCtx jwt token from context
func ParseFromCtx(ctx context.Context) (*Claims, error) {
	return defaultAPI.ParseFromCtx(ctx)
}
