package handlers

import (
	"strconv"
	"time"

	"github.com/cowry-network/go/clients/federation"
	"github.com/cowry-network/go/clients/stellartoml"
	"github.com/cowry-network/go/services/compliance/internal/config"
	"github.com/cowry-network/go/services/compliance/internal/crypto"
	"github.com/cowry-network/go/services/compliance/internal/db"
	"github.com/cowry-network/go/support/http"
)

// RequestHandler implements compliance server request handlers
type RequestHandler struct {
	Config                  *config.Config                 `inject:""`
	Client                  http.SimpleHTTPClientInterface `inject:""`
	Database                db.Database                    `inject:""`
	SignatureSignerVerifier crypto.SignerVerifierInterface `inject:""`
	StellarTomlResolver     stellartoml.ClientInterface    `inject:""`
	FederationResolver      federation.ClientInterface     `inject:""`
	NonceGenerator          NonceGeneratorInterface        `inject:""`
}

type NonceGeneratorInterface interface {
	Generate() string
}

type NonceGenerator struct{}

func (n *NonceGenerator) Generate() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

type TestNonceGenerator struct{}

func (n *TestNonceGenerator) Generate() string {
	return "nonce"
}
