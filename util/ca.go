package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type RootCertificateAuthority struct {
	template    *x509.Certificate
	privateKey  *rsa.PrivateKey
	certificate []byte
	issuedCerts map[string][][]byte
}

func NewRootCertificateAuthority() (*RootCertificateAuthority, error) {
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(int64(RandInt10000())),
		Subject: pkix.Name{
			Organization:  []string{"DIVINE-DRAGON"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"1337"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("can't generate a CA RSA key pair: %v", err)
	}
	caCert, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("can't create a CA certificate: %v", err)
	}
	rca := RootCertificateAuthority{
		template:    caTemplate,
		privateKey:  caPrivateKey,
		certificate: caCert,
	}
	rca.issuedCerts = make(map[string][][]byte)
	return &rca, nil
}

func (rca *RootCertificateAuthority) CreateTLSCert(host string) error {
	tlsCertTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(int64(RandInt10000())),
		Subject: pkix.Name{
			Organization:  []string{"DIVINE-DRAGON"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"1337"},
		},
		IPAddresses:  []net.IP{net.ParseIP(host)},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	tlsPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("can't generate a TLS RSA key pair: %v", err)
	}
	tlsCert, err := x509.CreateCertificate(rand.Reader, tlsCertTemplate, rca.template, &tlsPrivateKey.PublicKey, rca.privateKey)
	if err != nil {
		return err
	}
	pkcs1Key := x509.MarshalPKCS1PrivateKey(tlsPrivateKey)
	rca.issuedCerts[host] = [][]byte{pkcs1Key, tlsCert}
	return nil
}

func (rca *RootCertificateAuthority) DumpAll() error {
	rootCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rca.certificate,
	})
	err := WriteToFile("data/c2/root.crt", string(rootCert))
	if err != nil {
		return fmt.Errorf("can't write to a file: %v", err)
	}
	rootKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rca.privateKey),
	})
	err = WriteToFile("data/c2/root.key", string(rootKey))
	if err != nil {
		return fmt.Errorf("can't write to a file: %v", err)
	}
	for host, issued := range rca.issuedCerts {
		issuedCert := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: issued[1],
		})
		err := WriteToFile("data/c2/"+host+".crt", string(issuedCert))
		if err != nil {
			return fmt.Errorf("can't write to a file: %v", err)
		}
		issuedKey := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: issued[0],
		})
		err = WriteToFile("data/c2/"+host+".key", string(issuedKey))
		if err != nil {
			return fmt.Errorf("can't write to a file: %v", err)
		}
	}
	return nil
}
