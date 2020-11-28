package scanner

import (
	"log"
	"net"
)

// GetLocalIPs returns all IPs on the local host
func getLocalIPs() []net.IP {
	ips := make([]net.IP, 0, 12)

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("error getting host network interfaces", err)
	}

	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal("error getting host addresses", err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			ips = append(ips, ip)
		}
	}

	return ips
}
