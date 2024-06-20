package dblock

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func LookupIP(domain string, dns string) ([]string, error) {

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {

			ip := net.ParseIP(dns)

			if ip.To4() == nil {
				dialer := net.Dialer{
					Timeout:   5e9, // 5 seconds
					KeepAlive: 1e9, // 1 second
				}
				return dialer.DialContext(ctx, "udp", fmt.Sprintf("[%s]:53", dns))

			} else {
				return net.Dial("udp", dns+":53")
			}
		},
	}

	// perform a DNS lookup using the custom resolver
	ips, err := resolver.LookupIPAddr(context.Background(), domain)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, ErrorNoIPHost
	}

	ip := make([]string, 0)
	for i := 0; i < len(ips); i++ {
		t := ips[i].IP.String()

		t = strings.TrimSpace(t)
		t = strings.Trim(t, ".")

		if len(t) > 0 {
			ip = append(ip, t)
		}
	}

	if len(ip) == 0 {
		return nil, ErrorNoIPHost
	}
	return ip, nil
}
func LookupHost(domain string, dns string) ([]string, error) {

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {

			ip := net.ParseIP(dns)

			if ip.To4() == nil {
				dialer := net.Dialer{
					Timeout:   5e9, // 5 seconds
					KeepAlive: 1e9, // 1 second
				}
				return dialer.DialContext(ctx, "udp", fmt.Sprintf("[%s]:53", dns))

			} else {
				return net.Dial("udp", dns+":53")
			}
		},
	}

	// perform a DNS lookup using the custom resolver
	hosts, err := resolver.LookupHost(context.Background(), domain)
	if err != nil {
		if strings.Contains(err.Error(), "connect: network is unreachable") {
			log.Fatal("connect: network is unreachable")
		}
		return nil, ErrorNoIPHost
	}

	ho := make([]string, 0)
	for i := 0; i < len(hosts); i++ {
		t := hosts[i]

		t = strings.TrimSpace(t)
		t = strings.Trim(t, ".")

		if len(t) > 0 {
			ho = append(ho, t)
		}
	}

	if len(ho) == 0 {
		return nil, ErrorNoIPHost
	}

	return ho, nil
}
func LookupNS(domain string, dns string, trackResponseTimes bool) ([]string, error) {

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {

			ip := net.ParseIP(dns)

			if ip.To4() == nil {
				dialer := net.Dialer{
					Timeout:   5e9, // 5 seconds
					KeepAlive: 1e9, // 1 second
				}
				return dialer.DialContext(ctx, "udp", fmt.Sprintf("[%s]:53", dns))

			} else {
				return net.Dial("udp", dns+":53")
			}
		},
	}
	var err error
	var hosts []*net.NS

	if trackResponseTimes {
		f, err := os.OpenFile("responsetimes.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()

		t := time.Now()
		// perform a DNS lookup using the custom resolver
		hosts, err = resolver.LookupNS(context.Background(), domain)
		dur := time.Since(t)
		respStr := fmt.Sprintf("%s;%d;\n", dns, dur.Milliseconds())
		f.WriteString(respStr)
	} else {
		hosts, err = resolver.LookupNS(context.Background(), domain)
	}

	if err != nil {
		if strings.Contains(err.Error(), "connect: network is unreachable") {
			log.Fatal("connect: network is unreachable")
		}

		if strings.Contains(err.Error(), "no such host") {
			return nil, ErrorNoIPHost
		}

		// RATE LIMITING?
		// return nil, ErrorNoIPHost
		return nil, err
	}

	ho := make([]string, 0)
	for i := 0; i < len(hosts); i++ {
		t := hosts[i].Host

		t = strings.TrimSpace(t)
		t = strings.Trim(t, ".")

		if len(t) > 0 {
			ho = append(ho, t)
		}
	}

	if len(ho) == 0 {
		return nil, ErrorNoIPHost
	}

	return ho, nil
}
