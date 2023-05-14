package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	Directory = "/tmp/k8s-webhook-server/serving-certs"
)

func GenerateIfNotExists(subject string) error {
	if _, err := os.Stat(filepath.Join(Directory, "tls.crt")); err == nil {
		return nil
	}

	// 인증서와 개인 키 생성
	caCert, caKey, err := generateCertificateAuthority(subject)
	if err != nil {
		fmt.Println("Failed to generate CA certificate:", err)
		return err
	}

	// 인증서를 파일로 저장
	caCertFile, err := os.Create(filepath.Join(Directory, "tls.crt"))
	if err != nil {
		fmt.Println("Failed to create CA certificate file:", err)
		return err
	}
	defer caCertFile.Close()
	pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// 개인 키를 파일로 저장
	caKeyFile, err := os.Create(filepath.Join(Directory, "tls.key"))
	if err != nil {
		fmt.Println("Failed to create CA key file:", err)
		return err
	}
	defer caKeyFile.Close()
	pem.Encode(caKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})

	fmt.Println("Self-signed certificate and key generated successfully.")
	return nil
}

func generateCertificateAuthority(subject string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// 개인 키 생성
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 인증서 템플릿 생성
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: subject},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10년간 유효
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 인증서 생성
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}
