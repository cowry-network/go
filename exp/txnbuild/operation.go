package txnbuild

import (
	"github.com/cowry-network/go/xdr"
)

// Operation represents the operation types of the Stellar network.
type Operation interface {
	BuildXDR() (xdr.Operation, error)
}
