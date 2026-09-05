package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

type NetworkInfo struct {
	Interface string
	IPv4      string
	IPv6      string
	Gateway   string
	DNS       []string
	Online    bool
	Latency   float64
}

func GetNetworkInfo() NetworkInfo {
	info := NetworkInfo{}

	var interfaceIP string

	info.Gateway, interfaceIP = GetDefaultGatewayAndInterfaceIP()

	if interfaceIP == "" {
		return info
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		foundInterface := false

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.To4() != nil && ip.String() == interfaceIP {
				info.Interface = iface.Name
				info.IPv4 = ip.String()
				foundInterface = true
				break
			}
		}

		if !foundInterface {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.To4() == nil {
				info.IPv6 = ip.String()
				break
			}
		}

		info.DNS = GetDNS(iface.Name)
		temp_online, temp_lat := CheckOnline()
		info.Online = temp_online
		info.Latency = temp_lat

		return info
	}

	return info
}

func GetDefaultGatewayAndInterfaceIP() (string, string) {
	cmd := exec.Command("route", "print", "0.0.0.0")
	output, err := cmd.Output()

	if err != nil {
		return "", ""
	}

	for line := range strings.SplitSeq(string(output), "\n") {
		fields := strings.Fields(line)

		if len(fields) >= 4 &&
			fields[0] == "0.0.0.0" &&
			fields[1] == "0.0.0.0" {
			return fields[2], fields[3]
		}
	}
	return "", ""
}

func GetDNS(interfaceName string) []string {
	var dns []string

	cmd := exec.Command("ipconfig", "/all")

	output, err := cmd.Output()
	if err != nil {
		return dns
	}

	lines := strings.Split(string(output), "\n")

	inTargetInterface := false

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		if !strings.HasPrefix(rawLine, " ") &&
			strings.HasSuffix(line, ":") {
			inTargetInterface = strings.Contains(line, interfaceName+":")
			continue
		}

		if !inTargetInterface {
			continue
		}

		if strings.HasPrefix(line, "DNS Servers") {
			parts := strings.SplitN(line, ":", 2)

			if len(parts) != 2 {
				continue
			}

			server := strings.TrimSpace(parts[1])

			if server != "" {
				dns = append(dns, server)
			}
		}
	}

	return dns
}

func CheckOnline() (bool, float64) {
	start := time.Now()
	conn, error := net.DialTimeout(
		"tcp",
		"8.8.8.8:53",
		2*time.Second,
	)

	if error != nil {
		return false, -1
	}

	conn.Close()
	latency := float64(time.Since(start).Microseconds()) / 1000

	return true, latency
}

func (n NetworkInfo) Display() {
	fmt.Println()
	fmt.Println("╭──────────────────────────────────────────────╮")
	fmt.Println("│                 NETLYZER                     │")
	fmt.Println("╰──────────────────────────────────────────────╯")

	fmt.Println()
	fmt.Println("╭─ Network ────────────────────────────────────╮")
	fmt.Printf("│ Interface : %-32s │\n", n.Interface)
	fmt.Printf("│ IPv4      : %-32s │\n", n.IPv4)
	fmt.Printf("│ IPv6      : %-32s │\n", n.IPv6)
	fmt.Printf("│ Gateway   : %-32s │\n", n.Gateway)

	dns := "None"
	if len(n.DNS) > 0 {
		dns = strings.Join(n.DNS, ", ")
	}

	fmt.Printf("│ DNS       : %-32s │\n", dns)

	status := "OFFLINE"
	if n.Online {
		status = "ONLINE"
	}

	fmt.Printf("│ Status    : %-32s │\n", status)

	fmt.Println("╰──────────────────────────────────────────────╯")

	fmt.Println()
	fmt.Println("╭─ Connection ─────────────────────────────────╮")

	if n.Latency >= 0 {
		fmt.Printf("│ Latency   : %-35s│\n", fmt.Sprintf("%.2f ms", n.Latency))
	} else {
		fmt.Printf("│ Latency   : %-32s │\n", "N/A")
	}

	fmt.Println("╰──────────────────────────────────────────────╯")
	fmt.Println()
}
