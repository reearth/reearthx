package authserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"gopkg.in/square/go-jose.v2"
)

func initKeys(keyBytes, certBytes []byte) (*rsa.PrivateKey, *jose.SigningKey, *jose.JSONWebKeySet, error) {
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, nil, nil, fmt.Errorf("failed to decode the key bytes")
	}
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse the private key bytes: %w\n", err)
	}

	var certActualBytes []byte
	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		certActualBytes = certBytes // backwards compatibility
	} else {
		certActualBytes = certBlock.Bytes
	}

	var cert *x509.Certificate
	cert, err = x509.ParseCertificate(certActualBytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse the cert bytes: %w\n", err)
	}

	keyID := "RE01"
	sk := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       jose.JSONWebKey{Key: key, Use: "sig", Algorithm: string(jose.RS256), KeyID: keyID, Certificates: []*x509.Certificate{cert}},
	}

	return key, &sk, &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{Key: key.Public(), Use: "sig", Algorithm: string(jose.RS256), KeyID: keyID, Certificates: []*x509.Certificate{cert}},
		},
	}, nil
}

func generateCert(name pkix.Name) (keyPem, certPem []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("failed to generate key: %w\n", err)
		return
	}

	keyPem = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      name,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(100, 0, 0),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, key.Public(), key)
	if err != nil {
		err = fmt.Errorf("failed to create the cert: %w\n", err)
	}

	certPem = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	return
}
