//
// Copyright 2017 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package check

import (
	"os"
	"time"

	"bytes"
	"encoding/gob"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"sync"
)

const (
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
)

// Pinger represents the facility to send out a single ping packet and provides the response via the returned channel.
type Pinger interface {
	Ping(seq int) <-chan *PingResponse
}

// PingerFactorySpec specifies function specification to use
// when creating a Pinger.
// ipVersion is either "v4" or "v6"
type PingerFactorySpec func(identifier string, remoteAddr string, ipVersion string) (Pinger, error)

// PingerFactory creates and returns a new pinger with a specified address
// It then delegates th pingerImpl private function to wrap the pinger
// with the Timeouts, Counts, Statistics, and Run methods.
// ipVersion is either "v4" or "v6"
type PingResponse struct {
	Seq int
	Rtt time.Duration
	Err error
}

type pingRoutingKey struct {
	identifier string
	id         int
}

type pingRouter struct {
	mu sync.Mutex
	// key is "network"
	packetConns map[string]*icmp.PacketConn
	consumers   map[pingRoutingKey]chan *PingResponse
}

var PingerFactory PingerFactorySpec = func(identifier string, remoteAddr string, ipVersion string) (Pinger, error) {
	privileged := os.Geteuid() == 0
	if !privileged {
		switch runtime.GOOS {
		case "darwin":
		case "linux":
			log.Debug("Using non-privileged, UDP ping")
		default:
			return nil, fmt.Errorf("Non-privileged ping is not supported on %v", runtime.GOOS)
		}
	}

	var network string
	switch {
	case privileged && ipVersion == "v4":
		network = "ip4:imcp"
	case privileged && ipVersion == "v6":
		network = "ip6:ipv6-icmp"
	case !privileged && ipVersion == "v4":
		network = "udp4"
	case !privileged && ipVersion == "v6":
		network = "udp6"
	}

	log.WithFields(log.Fields{
		"identifier": identifier,
		"remoteAddr": remoteAddr,
		"ipVersion":  ipVersion,
		"privileged": privileged,
		"network":    network,
	}).Debug("Factory creating pinger")

	return NewPinger(identifier, network, remoteAddr)
}

var (
	defaultPingRouter  *pingRouter
	pingNetworkToProto = map[string]int{
		"udp4":          ProtocolICMP,
		"udp6":          ProtocolIPv6ICMP,
		"ip4:imcp":      ProtocolICMP,
		"ip6:ipv6-icmp": ProtocolIPv6ICMP,
	}
)

func init() {
	defaultPingRouter = &pingRouter{
		packetConns: make(map[string]*icmp.PacketConn, 4),
		consumers:   make(map[pingRoutingKey]chan *PingResponse),
	}
}

// getPacketConn ensures that a PacketConn is created if needed for the requested IMCP network type and will also
// ensure a receiver is running for that PacketConn
func (pr *pingRouter) getPacketConn(network string, bindAddr string) (*icmp.PacketConn, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	packetConn, ok := pr.packetConns[network]
	if ok {
		return packetConn, nil
	}

	packetConn, err := icmp.ListenPacket(network, bindAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create ICMP listener")
	}
	pr.packetConns[network] = packetConn

	go pr.receive(network, packetConn)

	return packetConn, nil
}

func (pr *pingRouter) receive(network string, packetConn *icmp.PacketConn) {
	log.WithFields(log.Fields{"network": network}).Debug("Starting ICMP receiver")
	defer log.WithFields(log.Fields{"network": network}).Debug("Stopping ICMP receiver")
	buffer := make([]byte, 1500)

recvLoop:
	for {
		n, peerAddr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "network": network}).Warn("Reading from ICMP connection")

			pr.mu.Lock()
			packetConn.Close()
			delete(pr.packetConns, network)
			pr.mu.Unlock()
			return
		}

		proto := pingNetworkToProto[network]
		m, err := icmp.ParseMessage(proto, buffer[:n])
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err,
				"network":  network,
				"peerAddr": peerAddr,
			}).Warn("Failed to parse ICMP message")
			continue
		}

		switch m.Type {
		case ipv4.ICMPTypeEchoReply:
		case ipv6.ICMPTypeEchoReply:
		case ipv4.ICMPTypeDestinationUnreachable:
		case ipv6.ICMPTypeDestinationUnreachable:
			break
		default:
			log.WithFields(log.Fields{
				"type":     m.Type,
				"peerAddr": peerAddr,
			}).Warn("Received non echo reply")
			continue recvLoop
		}

		switch pkt := m.Body.(type) {
		case *icmp.Echo:
			id := pkt.ID
			seq := pkt.Seq
			rbuf := bytes.NewBuffer(pkt.Data)
			dec := gob.NewDecoder(rbuf)
			var payload echoPayload
			err := dec.Decode(&payload)
			if err != nil {
				log.WithFields(log.Fields{
					"id":   id,
					"seq":  seq,
					"data": pkt.Data,
				}).Warn("Failed to decode echo reply payload")
				continue recvLoop
			}

			pingResponse := &PingResponse{
				Seq: seq,
				Rtt: time.Since(payload.Sent),
			}

			pr.routeResponse(payload.Identifier, id, seq, pkt.Data, pingResponse)

		case *icmp.DstUnreach:
			log.WithFields(log.Fields{
				"peerAddr":   peerAddr,
				"data":       pkt.Data,
				"extensions": pkt.Extensions,
			}).Debug("Received DstUnreach")

			/*
				TODO decipher this type of example response
				packet id=27289 identifier=ch00000 remoteAddr=127.0.0.2:0 seq=10
				data=[69 0 107 1 1 111 0 0 64 17 0 0 127 0 0 1 127 0 0 1 209 0 0 67 1 87 0 0]
			*/

			continue recvLoop

		default:
			log.WithFields(log.Fields{
				"body":     m.Body,
				"peerAddr": peerAddr,
			}).Warn("Received non echo body")

			continue recvLoop
		}

	}
}

func (pr *pingRouter) routeResponse(identifier string, id int, seq int, data []byte, pingResponse *PingResponse) {
	pr.mu.Lock()
	ch, ok := pr.consumers[pingRoutingKey{identifier: identifier, id: id}]
	if !ok {
		log.WithFields(log.Fields{
			"identifier": identifier,
			"id":         id,
			"seq":        seq,
			"data":       data,
		}).Warn("Unable to find ping routing consumer")
		return
	}
	pr.mu.Unlock()

	ch <- pingResponse

}

func (pr *pingRouter) register(identifier string, id int, responses chan *PingResponse) {
	pr.mu.Lock()

	pr.consumers[pingRoutingKey{identifier: identifier, id: id}] = responses

	pr.mu.Unlock()
}

func (pr *pingRouter) deregister(identifier string, id int) {
	pr.mu.Lock()

	delete(pr.consumers, pingRoutingKey{identifier: identifier, id: id})

	pr.mu.Unlock()
}

type pingerBase struct {
	packetConn *icmp.PacketConn
	id         int
	identifier string
	remoteAddr net.Addr
	responses  chan *PingResponse
}

func newPingerBase(identifier string, packetConn *icmp.PacketConn, remoteAddr net.Addr) *pingerBase {
	id := rand.Intn(0xffff)
	responses := make(chan *PingResponse, 1)

	defaultPingRouter.register(identifier, id, responses)

	return &pingerBase{
		packetConn: packetConn,
		id:         id,
		identifier: identifier,
		remoteAddr: remoteAddr,
		responses:  responses,
	}
}

type pingerV4 struct {
	pingerBase
}

type pingerV6 struct {
	pingerBase
}

func NewPinger(identifier string, network string, remoteAddr string) (Pinger, error) {
	var bindAddr string
	switch network {
	case "ip4:imcp":
	case "udp4":
		bindAddr = "0.0.0.0"
		packetConn, err := defaultPingRouter.getPacketConn(network, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip4", network, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV4{pingerBase: *newPingerBase(identifier, packetConn, addr)}, nil
	case "ip6:ipv6-icmp":
	case "udp6":
		bindAddr = "::"
		packetConn, err := defaultPingRouter.getPacketConn(network, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip6", network, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV6{pingerBase: *newPingerBase(identifier, packetConn, addr)}, nil
	}

	return nil, fmt.Errorf("Unsupported network type: %v", network)
}

func resolvePingAddr(addrNetwork string, imcpNetwork string, remoteAddr string) (net.Addr, error) {
	ipAddr, err := net.ResolveIPAddr(addrNetwork, remoteAddr)
	if err != nil {
		return nil, err
	}

	var pingAddr net.Addr
	if strings.HasPrefix(imcpNetwork, "udp") {
		pingAddr = &net.UDPAddr{IP: ipAddr.IP, Zone: ipAddr.Zone}
	} else {
		pingAddr = ipAddr
	}

	return pingAddr, nil
}

type echoPayload struct {
	Identifier string
	Sent       time.Time
}

func (p *pingerBase) ping(seq int, messageType icmp.Type) <-chan *PingResponse {
	payload := echoPayload{
		Identifier: p.identifier,
		Sent:       time.Now(),
	}
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(payload)
	if err != nil {
		p.responses <- &PingResponse{Err: err}
		return p.responses
	}

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq,
			Data: buffer.Bytes(),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		p.responses <- &PingResponse{Err: err}
		return p.responses
	}

	log.WithFields(log.Fields{
		"identifier": p.identifier,
		"id":         p.id,
		"seq":        seq,
		"remoteAddr": p.remoteAddr,
	}).Debug("Sending ping packet")
	if _, err = p.packetConn.WriteTo(wb, p.remoteAddr); err != nil {
		p.responses <- &PingResponse{Err: err}
		return p.responses
	}

	return p.responses
}

func (p *pingerV4) Ping(seq int) <-chan *PingResponse {
	return p.ping(seq, ipv4.ICMPTypeEcho)
}

func (p *pingerV6) Ping(seq int) <-chan *PingResponse {
	return p.ping(seq, ipv6.ICMPTypeEchoRequest)
}
