package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type NetworkInfo struct {
	Interface string
	IPv4      string
	IPv6      string
	Gateway   string
	DNS       []string
	Online    bool
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

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.String() != interfaceIP {
				continue
			}

			if ip.To4() == nil {
				continue
			}

			info.Interface = iface.Name
			info.IPv4 = ip.String()
			info.Online = true

			return info
		}
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

func (n NetworkInfo) Display() {
	fmt.Println(n)
	fmt.Println("Interface:", n.Interface)
	fmt.Println("IPv4:", n.IPv4)
	fmt.Println("IPv6:", n.IPv6)
	fmt.Println("Gateway:", n.Gateway)
}
