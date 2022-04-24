package server

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/najibulloShapoatov/server-core/monitoring/log"
	"io/ioutil"
	"math/big"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// Manager interface is required by the server to be used as certificate provider
type Manager interface {
	GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error)
	GetCertificateFiles() (string, string)
	TLSConfig() *tls.Config
}

// externalCertManager returns the HTTPS certificate defined from system files
type externalCertManager struct {
	cert     *tls.Certificate
	certFile string
	keyFile  string
}

func newExternalCertificate(certFile, keyFile string) (m Manager) {
	return &externalCertManager{certFile: certFile, keyFile: keyFile}
}
func (v *externalCertManager) TLSConfig() *tls.Config                { return nil }
func (v *externalCertManager) GetCertificateFiles() (string, string) { return v.certFile, v.keyFile }
func (v *externalCertManager) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if v.cert != nil {
		return v.cert, nil
	}

	cert, err := tls.LoadX509KeyPair(v.certFile, v.keyFile)
	if err != nil {
		return nil, err
	}
	v.cert = &cert

	return &cert, err
}

// letsEncryptManager returns generates a certificate signed by Let's Encrypt on the first request it receives
// and generates the certificate for the domain requested by that first http request
type letsEncryptManager struct {
	manager *autocert.Manager
	dir     string
}

func newLetsEncryptManager(hosts []string) (m Manager) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("could not create temp folder: %s", err)
	}

	return &letsEncryptManager{
		manager: &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(hosts...),
			Cache:      autocert.DirCache(dir),
		},
		dir: dir,
	}
}
func (v *letsEncryptManager) TLSConfig() *tls.Config                { return v.manager.TLSConfig() }
func (v *letsEncryptManager) GetCertificateFiles() (string, string) { return "", "" }
func (v *letsEncryptManager) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return v.manager.GetCertificate(clientHello)
}

// selfSignManager will use GO to generate a certificate. This is mostly used during development
type selfSignManager struct {
	cert *tls.Certificate
}

func newSelfSignManager() (m Manager) {
	return &selfSignManager{}
}
func (v *selfSignManager) TLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: v.GetCertificate,
	}
}
func (v *selfSignManager) GetCertificateFiles() (string, string) { return "", "" }
func (v *selfSignManager) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if v.cert != nil {
		return v.cert, nil
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"ServerCore"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(fmt.Sprintf("localhost,%s", clientHello.ServerName), ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if false {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, err
	}
	certBuf := &bytes.Buffer{}
	if err := pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, err
	}

	keyBuf := &bytes.Buffer{}
	block, err := pemBlockForKey(priv)
	if err := pem.Encode(keyBuf, block); err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes())
	if err != nil {
		return nil, err
	}
	v.cert = &cert

	return &cert, err
}

func testKey(certFile, keyFile string, auto bool) (hasCert bool, err error) {
	c, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		if auto {
			log.Warnf("Provided certificate is invalid (%s). Using auto fetch", err)
			return false, nil
		} else {
			return false, fmt.Errorf("invalid certificate: %s", err)
		}
	} else {
		for _, raw := range c.Certificate {
			cert, _ := x509.ParseCertificate(raw)
			if cert.NotAfter.Before(time.Now()) {
				if auto {
					log.Warn("Certificate is expired. Using auto fetch")
					return false, nil
				} else {
					return false, fmt.Errorf("certificate is expired")
				}
			} else if cert.NotBefore.After(time.Now()) {
				if auto {
					log.Warn("Certificate is not yet valid. Using auto fetch")
					return false, nil
				} else {
					return false, fmt.Errorf("certificate is expired")
				}
			}
		}
	}
	return true, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, errors.New("no input")
	}
}
