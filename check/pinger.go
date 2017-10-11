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
	"fmt"
	log "github.com/sirupsen/logrus"
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
	ProtocolICMP              = 1
	ProtocolIPv6ICMP          = 58
	pingWriteDeadlineDuration = 1 * time.Minute
	pingReadDeadlineDuration  = 3 * time.Minute

	PingerIPv4 = "v4"
	PingerIPv6 = "v6"

	IcmpNetIP4  = "ip4:icmp"
	IcmpNetUDP4 = "udp4"
	IcmpNetIP6  = "ip6:ipv6-icmp"
	IcmpNetUDP6 = "udp6"

	pingLogPrefix = "pinger"

	GooglePingLimit = 64
)

// Pinger represents the facility to send out a single ping packet and provides the response via the returned channel.
type Pinger interface {
	Ping(seq int) <-chan *PingResponse
	Close()
}

// PingerFactorySpec specifies function specification to use
// when creating a Pinger.
// ipVersion is "v4" or "v6" or "" to auto-interpret
type PingerFactorySpec func(identifier string, remoteAddr string, ipVersion string) (Pinger, error)

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

// PingerFactory creates and returns a new pinger that is initialized to a standard implementation, but can be
// swapped out for unit testing, etc.
var PingerFactory PingerFactorySpec = func(identifier string, remoteAddr string, ipVersion string) (Pinger, error) {
	privileged := os.Geteuid() == 0
	if !privileged {
		switch runtime.GOOS {
		case "darwin", "linux":
			// supported
		default:
			return nil, fmt.Errorf("Non-privileged ping is not supported on %v", runtime.GOOS)
		}
	}

	if ipVersion == "" {
		ipAddr, err := net.ResolveIPAddr("ip", remoteAddr)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to guess at IPv4/v6 address type")
		}

		if ipAddr.IP.To4() != nil {
			ipVersion = PingerIPv4
		} else {
			ipVersion = PingerIPv6
		}
	}

	var network string
	switch {
	case privileged && ipVersion == PingerIPv4:
		network = IcmpNetIP4
	case privileged && ipVersion == PingerIPv6:
		network = IcmpNetIP6
	case !privileged && ipVersion == PingerIPv4:
		network = IcmpNetUDP4
	case !privileged && ipVersion == PingerIPv6:
		network = IcmpNetUDP6
	}

	log.WithFields(log.Fields{
		"prefix":     pingLogPrefix,
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
		IcmpNetUDP4: ProtocolICMP,
		IcmpNetUDP6: ProtocolIPv6ICMP,
		IcmpNetIP4:  ProtocolICMP,
		IcmpNetIP6:  ProtocolIPv6ICMP,
	}
)

func init() {
	defaultPingRouter = &pingRouter{
		packetConns: make(map[string]*icmp.PacketConn, 4),
		consumers:   make(map[pingRoutingKey]chan *PingResponse),
	}
}

// getPacketConn ensures that a PacketConn is created if needed for the requested ICMP network type and will also
// ensure a receiver is running for that PacketConn
func (pr *pingRouter) getPacketConn(network string, bindAddr string) (*icmp.PacketConn, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	packetConn, ok := pr.packetConns[network]
	if ok {
		return packetConn, nil
	}

	log.WithFields(log.Fields{
		"prefix":   pingLogPrefix,
		"network":  network,
		"bindAddr": bindAddr,
	}).Debug("Creating ICMP packet connection")

	packetConn, err := icmp.ListenPacket(network, bindAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create ICMP listener")
	}
	pr.packetConns[network] = packetConn

	go pr.receive(network, packetConn)

	return packetConn, nil
}

// receive runs in a go routine and processes responses from requests initiated in pingerBase.ping
func (pr *pingRouter) receive(network string, packetConn *icmp.PacketConn) {
	log.WithFields(log.Fields{
		"prefix":  pingLogPrefix,
		"network": network,
	}).Debug("Starting ICMP receiver")
	defer log.WithFields(log.Fields{
		"prefix":  pingLogPrefix,
		"network": network,
	}).Debug("Stopping ICMP receiver")

	buffer := make([]byte, 1500)

recvLoop:
	for {
		packetConn.SetReadDeadline(time.Now().Add(pingReadDeadlineDuration))
		n, peerAddr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					continue recvLoop
				}
			}

			log.WithFields(log.Fields{
				"prefix":  pingLogPrefix,
				"err":     err,
				"network": network,
			}).Warn("Reading from ICMP connection")

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
				"prefix":   pingLogPrefix,
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
		default:
			log.WithFields(log.Fields{
				"prefix":   pingLogPrefix,
				"type":     m.Type,
				"peerAddr": peerAddr,
			}).Debug("Received non echo reply")
			continue recvLoop
		}

		switch pkt := m.Body.(type) {
		case *icmp.Echo:
			id := pkt.ID
			seq := pkt.Seq
			rbuf := bytes.NewBuffer(pkt.Data)
			// This could be a cross-chatter echo reply, so need to constrain the amount of string reading below
			if rbuf.Len() > GooglePingLimit {
				rbuf.Truncate(GooglePingLimit)
			}

			identifier, err := rbuf.ReadString(0)
			if err != nil {
				log.WithFields(log.Fields{
					"prefix": pingLogPrefix,
					"id":     id,
					"seq":    seq,
					"data":   pkt.Data,
				}).Debug("Failed to decode identifier from echo reply payload")
				continue recvLoop
			}
			// trim off the delimiter
			identifier = identifier[:len(identifier)-1]

			var sent time.Time
			err = sent.UnmarshalBinary(rbuf.Bytes())
			if err != nil {
				log.WithFields(log.Fields{
					"prefix": pingLogPrefix,
					"id":     id,
					"seq":    seq,
					"data":   pkt.Data,
				}).Debug("Failed to decode sent time from echo reply payload")
				continue recvLoop
			}

			pingResponse := &PingResponse{
				Seq: seq,
				Rtt: time.Since(sent),
			}

			pr.routeResponse(identifier, id, seq, pkt.Data, pingResponse)

		case *icmp.DstUnreach:
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
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
				"prefix":   pingLogPrefix,
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
			"prefix":     pingLogPrefix,
			"identifier": identifier,
			"id":         id,
			"seq":        seq,
			"data":       data,
		}).Debug("Unable to find ping routing consumer")
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

func (p *pingerBase) Close() {
	defaultPingRouter.deregister(p.identifier, p.id)
}

type pingerV4 struct {
	pingerBase
}

type pingerV6 struct {
	pingerBase
}

// NewPinger directly creates a new instance with the given details.
// icmpNetwork should be one of IcmpNetIP4, IcmpNetUDP4, IcmpNetIP6, IcmpNetUDP6
func NewPinger(identifier string, icmpNetwork string, remoteAddr string) (Pinger, error) {
	log.WithFields(log.Fields{
		"prefix":     pingLogPrefix,
		"identifier": identifier,
		"network":    icmpNetwork,
		"remoteAddr": remoteAddr,
	}).Debug("Creating new pinger")

	var bindAddr string
	switch icmpNetwork {
	case IcmpNetIP4, IcmpNetUDP4:
		bindAddr = "0.0.0.0"
		packetConn, err := defaultPingRouter.getPacketConn(icmpNetwork, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip4", icmpNetwork, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV4{pingerBase: *newPingerBase(identifier, packetConn, addr)}, nil
	case IcmpNetIP6, IcmpNetUDP6:
		bindAddr = "::"
		packetConn, err := defaultPingRouter.getPacketConn(icmpNetwork, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip6", icmpNetwork, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV6{pingerBase: *newPingerBase(identifier, packetConn, addr)}, nil
	}

	return nil, fmt.Errorf("Unsupported network type: %v", icmpNetwork)
}

func resolvePingAddr(addrNetwork string, icmpNetwork string, remoteAddr string) (net.Addr, error) {
	ipAddr, err := net.ResolveIPAddr(addrNetwork, remoteAddr)
	if err != nil {
		return nil, err
	}

	var pingAddr net.Addr
	if strings.HasPrefix(icmpNetwork, "udp") {
		pingAddr = &net.UDPAddr{IP: ipAddr.IP, Zone: ipAddr.Zone}
	} else {
		pingAddr = ipAddr
	}

	return pingAddr, nil
}

func (p *pingerBase) ping(seq int, messageType icmp.Type) <-chan *PingResponse {
	// google.com truncates any ping bodies greater than 64 bytes, so hand-encoding our two fields
	nowBytes, err := time.Now().MarshalBinary()
	if err != nil {
		p.responses <- &PingResponse{Err: err}
		return p.responses
	}
	var buffer bytes.Buffer
	buffer.WriteString(p.identifier)
	buffer.WriteByte(0)
	buffer.Write(nowBytes)

	wm := icmp.Message{
		Type: messageType,
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
		"prefix":     pingLogPrefix,
		"identifier": p.identifier,
		"id":         p.id,
		"seq":        seq,
		"remoteAddr": p.remoteAddr,
	}).Debug("Sending ping packet")

	p.packetConn.SetWriteDeadline(time.Now().Add(pingWriteDeadlineDuration))
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
