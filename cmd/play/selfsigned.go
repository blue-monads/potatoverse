package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
	rsaBits  = 2048
)

// fileExists checks if a file exists.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// generateSelfSignedCert generates a new self-signed certificate and private key
// if they don't already exist.
func generateSelfSignedCert() error {
	// Check if files already exist
	if fileExists(certFile) && fileExists(keyFile) {
		log.Printf("Found existing certificate (%s) and key (%s).", certFile, keyFile)
		return nil
	}

	log.Printf("Generating new self-signed certificate (%s) and key (%s)...", certFile, keyFile)

	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Local Dev Server"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0), // Valid for 1 year

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:        true, // We are self-signing

		BasicConstraintsValid: true,
		// Add localhost and 127.0.0.1 so browsers will trust it
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:    []string{"localhost"},
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// --- Save Certificate (cert.pem) ---
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}
	log.Printf("Successfully created %s", certFile)

	// --- Save Private Key (key.pem) ---
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}
	if err := keyOut.Close(); err != nil {
		return err
	}
	log.Printf("Successfully created %s", keyFile)

	return nil
}

func SelfSignedServer() {
	// 1. Ensure we have a certificate and key
	if err := generateSelfSignedCert(); err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// 2. Set up the Gin router
	r := gin.Default()

	// 3. Define a simple route
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, world from a secure Gin server!")
	})

	// 4. Start the HTTPS server
	log.Println("Starting Gin server on https://localhost:8080 ...")
	// r.RunTLS will use the provided cert and key files
	if err := r.RunTLS(":8080", certFile, keyFile); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
