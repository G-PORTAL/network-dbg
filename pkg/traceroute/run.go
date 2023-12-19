package traceroute

import (
	"context"
	"fmt"
	"github.com/G-PORTAL/network-dbg/pkg/locations"
	"github.com/charmbracelet/log"
	"github.com/pixelbender/go-traceroute/traceroute"
	"net"
	"strings"
	"sync"
	"time"
)

type result struct {
	endpoint locations.LocationEndpoint
	reply    *traceroute.Reply
}

func Run(endpoints []locations.LocationEndpoint) {
	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	t := &traceroute.Tracer{
		Config: traceroute.Config{
			Delay:   50 * time.Millisecond,
			Timeout: 3 * time.Second,
			MaxHops: 30,
			Count:   1,

			Networks: []string{"ip4:icmp", "ip4:ip"},
		},
	}

	defer t.Close()

	results := make(chan result)

	for _, endpoint := range endpoints {
		go func(tracer *traceroute.Tracer, endpoint locations.LocationEndpoint) {
			defer wg.Done()

			ip, err := resolve(endpoint)
			if err != nil {
				log.Errorf("[%s]: Failed to resolve: %v", endpoint, err)
				return
			}

			err = tracer.Trace(context.Background(), ip, func(reply *traceroute.Reply) {
				results <- result{
					endpoint: endpoint,
					reply:    reply,
				}
			})

			if err != nil {
				log.Errorf("[%s]: Failed to trace: %v", endpoint, err)
			}

		}(t, endpoint)
	}

	log.Infof("Watiing for %d GPORTAL locations to finish...", len(endpoints))

	locationResults := make(map[locations.LocationEndpoint][]*traceroute.Reply)
	go func() {
		for {
			select {
			case r := <-results:
				locationResults[r.endpoint] = append(locationResults[r.endpoint], r.reply)
			default:
				continue
			}
		}
	}()

	wg.Wait()

	for endpoint, replies := range locationResults {
		fmt.Println()
		log.Infof("----------------------------------")
		log.Infof("GPORTAL Location: %s", endpoint.Short())
		log.Infof("----------------------------------")

		previousRTT := int64(0)
		for _, reply := range replies {

			hostname := ""
			if resolvedHostname, err := lookup(reply.IP); err == nil {
				hostname = " (" + resolvedHostname + ")"
			}

			msg := fmt.Sprintf("%v. %s%s (%s)", reply.Hops-1, reply.IP.String(), hostname, reply.RTT)
			if reply.RTT.Milliseconds()-previousRTT > 60 {
				log.Warnf(msg)
			} else {
				log.Infof(msg)
			}

			previousRTT = reply.RTT.Milliseconds()
		}
	}

	fmt.Println()
	log.Infof("Finished testing %d GPORTAL locations.", len(endpoints))
}

func resolve(endpoint locations.LocationEndpoint) (net.IP, error) {
	ips, err := net.LookupIP(string(endpoint))
	if err != nil {
		return nil, err
	}

	return ips[0], nil
}

func lookup(ip net.IP) (string, error) {
	names, err := net.LookupAddr(ip.String())
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(names[0], "."), nil
}
