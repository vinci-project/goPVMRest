package helpers

import (
	secp "github.com/vncsphere-foundation/secp256k1-go"
)

func PubkeyFromSeckey(privateKey []byte) []byte {
	//

	return secp.PubkeyFromSeckey([]byte(privateKey))
}
