package certificate

import (
	"crypto/tls"
	"fmt"
	"os"

	"software.sslmate.com/src/go-pkcs12"
)

// LoadCertificate carrega um certificado digital A1 (.pfx)
func LoadCertificate(certPath, password string) (tls.Certificate, error) {
	// Lê o arquivo do certificado
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Decodifica o certificado PKCS#12
	privateKey, certificate, err := pkcs12.Decode(certData, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to decode certificate: %w", err)
	}

	// Cria o certificado TLS
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certificate.Raw},
		PrivateKey:  privateKey,
		Leaf:        certificate,
	}

	return tlsCert, nil
}