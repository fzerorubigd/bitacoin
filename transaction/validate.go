package transaction

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"time"
)

func VerifySig(time time.Time, fromPubKey, toPubKey []byte, amount int, sig []byte) error {
	publicKey, err := x509.ParsePKCS1PublicKey(fromPubKey)
	if err != nil {
		return fmt.Errorf("parsing public key err: %s", err.Error())
	}

	hasher := sha256.New()
	_, err = fmt.Fprint(hasher, time, fromPubKey, toPubKey, amount)
	if err != nil {
		return fmt.Errorf("writing hash error in VerifySig err: %s", err.Error())
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), sig)
	if err != nil {
		return err
	}

	return nil
}

// TODO:
// 1.Does owner really had that money?
