package network

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

func GetUrl(url string) (serverHeader, requestUrlHeader, requestHostnameHeader string, err error) {
	resp, err := http.Get("https://" + url)
	if err != nil {
		return "", "", "", fmt.Errorf("%v", err)
	}

	serverHeader = resp.Header.Get("Server")
	if serverHeader == "" {
		serverHeader = "no value provided"
	}

	requestUrlHeader = resp.Request.URL.String()
	if requestUrlHeader == "" {
		requestUrlHeader = "no value provided"
	}

	requestHostnameHeader = resp.Request.URL.Hostname()
	if requestHostnameHeader == "" {
		requestHostnameHeader = "no value provided"
	}

	return serverHeader, requestUrlHeader, requestHostnameHeader, nil
}

func GetCertificates(url string) (certificates string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error: %v", err)
		return "", fmt.Errorf("%v", err)
	}

	serverName := resp.Request.URL.Hostname()
	if serverName == "" {
		serverName = "no value provided"
	}

	var builder strings.Builder

	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			for i, rawCert := range rawCerts {
				cert, err := x509.ParseCertificate(rawCert)
				if err != nil {
					return fmt.Errorf("failed to parse certificate: %v", err)
				}
				builder.WriteString(fmt.Sprintf("Certificate Chain #%d:\n", i))
				builder.WriteString(fmt.Sprintln("Subject:", cert.Subject.CommonName))
				builder.WriteString(fmt.Sprintln("Issuer:", cert.Issuer.CommonName))
				builder.WriteString(fmt.Sprintln("Valid From:", cert.NotBefore))
				builder.WriteString(fmt.Sprintln("Valid Until:", cert.NotAfter))
				builder.WriteString(fmt.Sprintln("Serial Number:", cert.SerialNumber))
				pemData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rawCert})
				builder.WriteString(fmt.Sprintf("PEM-encoded Certificate:\n%s\n", string(pemData)))
				builder.WriteString("-----\n")
			}

			return nil
		},
	}

	conn, err := tls.Dial("tcp", serverName+":443", config)
	if err != nil {
		log.Printf("error: %v", err)
		return "", fmt.Errorf("%v", err)
	}
	defer conn.Close()

	return builder.String(), nil

}

func GetDNS(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	// Filter out IPv6 addresses
	var ipv4Addresses []string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4Addresses = append(ipv4Addresses, ip.String())
		}
	}

	return ipv4Addresses, nil
}
