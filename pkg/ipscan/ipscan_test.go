package ipscan_test

import (
	"log"
	"testing"

	"reyzar.com/server-api/pkg/ipscan"
)

func TestIPScan(t *testing.T) {
	log.Println("result:", ipscan.IPScan("10.2.21.1-10.2.21.20"))
}
