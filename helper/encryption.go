package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func generateKeyPair(bits int) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return x509.MarshalPKCS1PrivateKey(privateKey),
		x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
		err
}

func GenerateWallet(dirPath string) ([]byte, []byte, error) {
	err := os.Mkdir(dirPath, os.ModePerm)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create wallet directory err: %w", err.Error())
	}

	priveKeyFile, err := os.Create(fmt.Sprintf("%s/private_key.txt", dirPath))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create wallet private_key file, err: %w", err.Error())
	}
	defer priveKeyFile.Close()

	publicKeyFile, err := os.Create(fmt.Sprintf("%s/public_key.txt", dirPath))
	if err != nil {
		return nil, nil, fmt.Errorf("could not create wallet public_key file, err: %w", err.Error())
	}
	defer publicKeyFile.Close()

	pr, pub, err := generateKeyPair(512)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key pair failed, err: %w", err.Error())
	}

	prPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: pr,
		},
	)
	_, err = priveKeyFile.Write(prPEM)
	if err != nil {
		return nil, nil, err
	}

	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pub,
		},
	)
	_, err = publicKeyFile.Write(pubPEM)
	if err != nil {
		return nil, nil, err
	}

	return pr, pub, nil
}

func ReadKeyFromPemFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	return block.Bytes, nil
}
