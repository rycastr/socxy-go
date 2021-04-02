package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"strings"
	"time"
)

type certificate struct {
	cert       []byte
	privateKey []byte
}

func NewCert(rsaBits int, duration int, host string) *certificate {
	pv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	pb := &pv.PublicKey
	keyUsage := x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	durationCert, err := time.ParseDuration(fmt.Sprintf("%vh", duration*24))
	if err != nil {
		log.Fatalf("Failed to parse duration of certificate: %v", err)
	}
	notAfter := notBefore.Add(durationCert)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    notBefore,
		NotAfter:     notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pb, pv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	pvBytes, err := x509.MarshalPKCS8PrivateKey(pv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}

	return &certificate{
		cert:       pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes}),
		privateKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pvBytes}),
	}
}

func (c *certificate) Certificate() []byte {
	return c.cert
}

func (c *certificate) PrivateKey() []byte {
	return c.privateKey
}
