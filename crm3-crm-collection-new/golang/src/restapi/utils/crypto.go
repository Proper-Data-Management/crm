package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"log"

	"golang.org/x/crypto/pkcs12"
)

func CryptoSignPKCS1v15(dataStr string, privateKeyFilePass string, dataToEncryptStr string) (string, string, error) {

	dataToEncrypt := []byte(dataToEncryptStr)

	// PRIVATE KEY
	data := []byte(dataStr)

	hashed := sha1.Sum(dataToEncrypt)

	privateKey, certificate, err := pkcs12.Decode(data, privateKeyFilePass)

	if err != nil {
		//log.Fatal(err)
		log.Println("error on pkcs12.Decode", err)
		return "", "", err
	}

	pv := privateKey.(*rsa.PrivateKey)
	//rsa.EncryptPKCS1v15()
	signature, err := rsa.SignPKCS1v15(rand.Reader, pv, crypto.SHA1, hashed[:])
	if err != nil {
		//log.Fatal(err)
		log.Println("error on rsa.SignPKCS1v15", err)
		return "", "", err
	}
	//log.Println("certificate", b64.StdEncoding.EncodeToString((certificate.Raw)))
	return string(signature), string(certificate.Raw), nil
}
