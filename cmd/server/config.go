package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultAddress = "127.0.0.1:19081"

type config struct {
	address  string
	database string
	selftest bool
}

func addressDefault() string {
	portText := strings.TrimSpace(os.Getenv("PORT"))
	if portText == "" {
		return defaultAddress
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return defaultAddress
	}
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
}

func validateAddress(address string, selftest bool) error {
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("监听地址必须为 host:port: %w", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("监听端口必须在 1 到 65535 之间")
	}
	if selftest {
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return errors.New("自检地址必须是明确的回环地址")
		}
	}
	return nil
}

func ensureDatabaseDirectory(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}
	directory := filepath.Dir(path)
	if directory == "." {
		return nil
	}
	return os.MkdirAll(directory, 0o750)
}
