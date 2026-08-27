package main

import "testing"

func TestValidateSelftestAddress(t *testing.T) {
	for _, address := range []string{"127.0.0.1:19081", "[::1]:19081", "localhost:19081"} {
		if err := validateAddress(address, true); err != nil {
			t.Fatalf("回环地址 %s 被拒绝: %v", address, err)
		}
	}
	for _, address := range []string{"0.0.0.0:19081", "192.0.2.10:19081", "127.0.0.1"} {
		if err := validateAddress(address, true); err == nil {
			t.Fatalf("含糊或非回环地址 %s 应被拒绝", address)
		}
	}
}

func TestAddressDefaultUsesValidPort(t *testing.T) {
	t.Setenv("PORT", "23119")
	if got := addressDefault(); got != "127.0.0.1:23119" {
		t.Fatalf("PORT 地址=%s", got)
	}
	t.Setenv("PORT", "invalid")
	if got := addressDefault(); got != defaultAddress {
		t.Fatalf("非法 PORT 应回退默认值，得到 %s", got)
	}
}
