package network

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/michaeljsaenz/traceroute"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func GetUrl(url string) (serverHeader, requestUrlHeader,
	requestHostnameHeader, requestResponseStatus string, err error) {
	resp, err := http.Get("https://" + url)
	if err != nil {
		return "", "", "", "", fmt.Errorf("%v", err)
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

	requestResponseStatus = resp.Status
	if requestResponseStatus == "" {
		requestResponseStatus = "no value provided"
	}

	return serverHeader, requestUrlHeader, requestHostnameHeader, requestResponseStatus, nil
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

func Ping(hostname string) (string, error) {
	ip, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		return "", fmt.Errorf("error resolving IP address: %w", err)
	}

	// create connection to send and receive ICMP messages
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return "", fmt.Errorf("error creating ICMP connection: %w", err)
	}
	defer conn.Close()

	// construct icmp messsage
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, // associate Echo Requests with Echo Replies
		},
	}

	// serialize the icmp message
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return "", fmt.Errorf("error marshaling ICMP message: %w", err)
	}

	sendTime := time.Now()

	// send the ICMP message to the listening connection
	_, err = conn.WriteTo(msgBytes, &net.UDPAddr{IP: net.ParseIP(ip.String())})
	if err != nil {
		return "", fmt.Errorf("error sending ICMP message: %w", err)
	}

	// deadline for receiving the icmp reply
	err = conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		return "", fmt.Errorf("error setting read deadline: %w", err)
	}

	// retry read icmp reply
	reply := make([]byte, 1500)
	numBytes, _, err := conn.ReadFrom(reply)
	if err != nil {
		return "", fmt.Errorf("error reading ICMP reply: %w", err)
	}

	// calculate RTT
	rtt := time.Since(sendTime)
	rttMilliseconds := rtt.Seconds() * 1000.0
	rttString := fmt.Sprintf("(Round Trip Time: %.3f ms)", rttMilliseconds)

	parsedReply, err := icmp.ParseMessage(1, reply[:numBytes])
	if err != nil {
		return "", fmt.Errorf("error on ICMP ParseMessage: %w", err)
	}

	switch parsedReply.Code {
	case 0:
		return fmt.Sprintf("Got reply from %s %s", hostname, rttString), nil
	case 3:
		return fmt.Sprintf("Host %s is unreachable", hostname), nil
	case 11:
		return fmt.Sprintf("Time exceeded attempting to reach host %s", hostname), nil
	default:
		return fmt.Sprintf("Host %s is unreachable", hostname), nil
	}

}

func Traceroute(hostname string) (hops string, err error) {
	opts := traceroute.TracerouteOptions{}
	opts.SetMaxHops(30)

	ipAddr, err := net.ResolveIPAddr("ip", hostname)
	if err != nil {
		return
	}

	c := make(chan traceroute.TracerouteHop)

	var builder strings.Builder
	hopStart := fmt.Sprintf("traceroute to %v (%v), %v hops max, %v byte packets\n", hostname, ipAddr, opts.MaxHops(), opts.PacketSize())
	builder.WriteString(hopStart)
	go func() {
		for {
			hop, ok := <-c
			if !ok {
				return
			}
			nextHop := printHop(hop)
			builder.WriteString(nextHop)
		}
	}()
	_, err = traceroute.Traceroute(hostname, &opts, c)
	if err != nil {
		return hopStart, fmt.Errorf("error: %w", err)
	}

	return builder.String(), nil
}

func printHop(hop traceroute.TracerouteHop) (nextHop string) {
	addr := fmt.Sprintf("%d.%d.%d.%d", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])

	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
	if hop.Success {
		nextHop = fmt.Sprintf("%-3d %v (%v) %s\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
		return nextHop
	} else {
		nextHop = fmt.Sprintf("%-3d *\n", hop.TTL)
		return nextHop
	}
}
