package transaction

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
)

func VerifySig(tnxRequest *TransactionRequest) error {
	publicKey, err := x509.ParsePKCS1PublicKey(tnxRequest.FromPubKey)
	if err != nil {
		return fmt.Errorf("parsing public key err: %s", err.Error())
	}

	hasher := sha256.New()
	_, err = fmt.Fprint(hasher, tnxRequest.Time, tnxRequest.FromPubKey,
		tnxRequest.ToPubKey, tnxRequest.FromPubKey, tnxRequest.Amount)
	if err != nil {
		return fmt.Errorf("writing hash error in VerifySig err: %s", err.Error())
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), tnxRequest.Signature)
	if err != nil {
		return err
	}

	return nil
}

// TODO:
// 1.Does owner really had that money?
// 2.Is the TNX id ok?
// 3.There should be only one coinbase transaction in block.
// 4.only one transaction free for each transaction
