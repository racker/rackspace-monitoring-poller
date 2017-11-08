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
)

const (
	ProtocolICMP              = 1
	ProtocolIPv6ICMP          = 58
	pingWriteDeadlineDuration = 1 * time.Minute
	PingReceiveBufferSize     = 1500

	PingerIPv4 = "v4"
	PingerIPv6 = "v6"

	IcmpNetIP4  = "ip4:icmp"
	IcmpNetUDP4 = "udp4"
	IcmpNetIP6  = "ip6:ipv6-icmp"
	IcmpNetUDP6 = "udp6"

	pingLogPrefix = "pinger"

	GooglePingLimit = 64

	// Ping by default pads the ICMP payload out to 64 bytes, but need to subtract 8 bytes for the ICMP header.
	// The "-s" option of ping in most man pages seems to unofficially document this.
	PadUpTo = 64 - 8
)

var (
	pingNetworkToProto = map[string]int{
		IcmpNetUDP4: ProtocolICMP,
		IcmpNetUDP6: ProtocolIPv6ICMP,
		IcmpNetIP4:  ProtocolICMP,
		IcmpNetIP6:  ProtocolIPv6ICMP,
	}

	// this is set during unit testing to enable very verbose logging of ping packet processing
	VerbosePinger = false
)

// Pinger represents the facility to send out a single ping packet and provides the response via the returned channel.
type Pinger interface {
	Ping(seq int, perPingDuration time.Duration) PingResponse
	Close()
}

// PingerFactorySpec specifies function specification to use
// when creating a Pinger.
// ipVersion is "v4" or "v6" or "" to auto-interpret
type PingerFactorySpec func(identifier string, remoteAddr string, ipVersion string) (Pinger, error)

type PingResponse struct {
	Seq     int
	Rtt     time.Duration
	Timeout bool
	Err     error
}

// pingerBase is the base implementation for both IPv4 and IPv6 flavors
type pingerBase struct {
	packetConn *icmp.PacketConn
	id         int
	identifier string
	remoteAddr net.Addr
	// proto is one of Protocol*Icmp from the icmp package
	proto int
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
			return nil, fmt.Errorf("non-privileged ping is not supported on %v", runtime.GOOS)
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

func createPacketConn(network string, bindAddr string) (*icmp.PacketConn, error) {

	packetConn, err := icmp.ListenPacket(network, bindAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create ICMP listener")
	}

	return packetConn, nil
}

func (p *pingerBase) ping(seq int, messageType icmp.Type, perPingDuration time.Duration) PingResponse {
	// google.com truncates any ping bodies greater than 64 bytes, so hand-encoding the timestamp
	nowBytes, err := time.Now().MarshalBinary()
	if err != nil {
		return PingResponse{Err: err}
	}
	// ...and the identifier, which is typically the check's ID
	var buffer bytes.Buffer
	buffer.WriteString(p.identifier)
	buffer.WriteByte(0)

	// time.Time marshaled preceded by the length of that
	buffer.WriteByte(byte(len(nowBytes)))
	buffer.Write(nowBytes)

	// a common convention seems to be padding out the ICMP packet with non-zero bytes
	var padByte byte = 1
	for buffer.Len() < PadUpTo {
		buffer.WriteByte(padByte)
		padByte++
	}

	wm := icmp.Message{
		Type: messageType,
		Code: 0,
		Body: &icmp.Echo{
			// The ICMP ID allows for filtering/correlation of the response
			ID: p.id,
			// ...and the response will also include the sequence we set here
			Seq:  seq,
			Data: buffer.Bytes(),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		return PingResponse{Err: err}
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
		return PingResponse{Err: err}
	}

	// now wait for the response packet matching our ID and remoteAddr
	return p.receive(perPingDuration)
}

//noinspection GoBoolExpressions
func (p *pingerBase) receive(perPingDuration time.Duration) PingResponse {

	buffer := make([]byte, PingReceiveBufferSize)

	// Use go's deadline concept to implement ping timeout handling. Since deadlines are an absolute time, we
	// can set it here once even though several foreign packets could get read and discarded in the recvLoop, below.
	err := p.packetConn.SetReadDeadline(time.Now().Add(perPingDuration))
	if err != nil {
		return PingResponse{Err: err}
	}

	// continuations here at recvLoop are due to observing ICMP response packets that are not ours
recvLoop:
	for {

		if VerbosePinger {
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
			}).Debug("Waiting for ICMP response")
		}

		// read a raw ICMP, but since ICMP is a broadcast protocol we don't know if it is ours yet...
		n, peerAddr, err := p.packetConn.ReadFrom(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					return PingResponse{Err: err, Timeout: true}
				}
			}

			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"err":        err,
			}).Warn("Reading from ICMP connection")

			p.packetConn.Close()
			return PingResponse{Err: err}
		}

		if VerbosePinger {
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"peerAddr":   peerAddr,
				"len":        n,
			}).Debug("Read packet")
		}

		// Is it even from the address we pinged? BTW net.Addr is an interface so we can do a plain old != on the
		// two addrs. Comparing String() rendering is cheesy but works reliably.
		if peerAddr.String() != p.remoteAddr.String() {
			continue recvLoop
		}

		m, err := icmp.ParseMessage(p.proto, buffer[:n])
		if err != nil {
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"err":        err,
				"peerAddr":   peerAddr,
			}).Warn("Failed to parse ICMP message")
			continue recvLoop
		}

		if VerbosePinger {
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"peerAddr":   peerAddr,
				"len":        n,
				"message":    m,
			}).Debug("Read ICMP message")
		}

		// Is the ICMP message type an expected reply type?
		switch m.Type {
		case ipv4.ICMPTypeEchoReply:
		case ipv6.ICMPTypeEchoReply:
		case ipv4.ICMPTypeDestinationUnreachable:
		case ipv6.ICMPTypeDestinationUnreachable:
		default:
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"type":       m.Type,
				"peerAddr":   peerAddr,
			}).Debug("Received non echo reply")
			continue recvLoop
		}

		// Cast and act upon the known/expected ICMP message body types
		switch pkt := m.Body.(type) {
		case *icmp.Echo:
			// An echo reply...just what we wanted

			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"ourId":      p.id,
				"type":       m.Type,
				"peerAddr":   peerAddr,
				"rxId":       pkt.ID,
				"rxSeq":      pkt.Seq,
			}).Debug("Received echo reply")

			// Is it our 16-bit random ID we included when sending out the packet
			id := pkt.ID
			if id != p.id {
				continue recvLoop
			}
			seq := pkt.Seq
			rbuf := bytes.NewBuffer(pkt.Data)
			// This could be a cross-chatter echo reply, so need to constrain the amount of string reading below
			if rbuf.Len() > GooglePingLimit {
				rbuf.Truncate(GooglePingLimit)
			}

			// We stored off the check's id within the ping request, so we'll extract and check
			identifier, err := rbuf.ReadString(0)
			if err != nil {
				log.WithFields(log.Fields{
					"prefix":     pingLogPrefix,
					"identifier": p.identifier,
					"id":         id,
					"seq":        seq,
					"data":       pkt.Data,
				}).Debug("Failed to decode identifier from echo reply payload")
				continue recvLoop
			}
			// trim off the delimiter
			identifier = identifier[:len(identifier)-1]

			// Next, we stored the timestamp in the ping request. Extract that and use it for round trip time (RTT) computation
			var sent time.Time
			timeLen, err := rbuf.ReadByte()
			if err != nil {
				log.WithFields(log.Fields{
					"prefix":     pingLogPrefix,
					"identifier": p.identifier,
					"id":         id,
					"seq":        seq,
					"data":       pkt.Data,
				}).Debug("Failed to decode time length from echo reply payload")
				continue recvLoop
			}
			err = sent.UnmarshalBinary(rbuf.Next(int(timeLen)))
			if err != nil {
				log.WithFields(log.Fields{
					"prefix":     pingLogPrefix,
					"identifier": p.identifier,
					"id":         id,
					"seq":        seq,
					"data":       pkt.Data,
				}).Debug("Failed to decode sent time from echo reply payload")
				continue recvLoop
			}

			// FINALLY...its ours, has a valid identifier, timestamp, so return the results

			pingResponse := PingResponse{
				Seq: seq,
				Rtt: time.Since(sent),
			}

			return pingResponse

		case *icmp.DstUnreach:
			// Eventually we may need to further process these, but can be treated as a missed/timed out response
			log.WithFields(log.Fields{
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
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
				"prefix":     pingLogPrefix,
				"identifier": p.identifier,
				"body":       m.Body,
				"peerAddr":   peerAddr,
			}).Warn("Received non echo body")

			continue recvLoop
		}

	}
}

func newPingerBase(identifier string, packetConn *icmp.PacketConn, remoteAddr net.Addr, network string) *pingerBase {
	id := rand.Intn(0xffff)

	log.WithFields(log.Fields{
		"prefix":     pingLogPrefix,
		"identifier": identifier,
		"remoteAddr": remoteAddr,
		"id":         id,
	}).Debug("Created new pinger")

	return &pingerBase{
		packetConn: packetConn,
		id:         id,
		identifier: identifier,
		remoteAddr: remoteAddr,
		proto:      pingNetworkToProto[network],
	}
}

func (p *pingerBase) Close() {
	p.packetConn.Close()
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

	var bindAddr string
	switch icmpNetwork {
	case IcmpNetIP4, IcmpNetUDP4:
		bindAddr = "0.0.0.0"
		packetConn, err := createPacketConn(icmpNetwork, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip4", icmpNetwork, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV4{pingerBase: *newPingerBase(identifier, packetConn, addr, icmpNetwork)}, nil
	case IcmpNetIP6, IcmpNetUDP6:
		bindAddr = "::"
		packetConn, err := createPacketConn(icmpNetwork, bindAddr)
		if err != nil {
			return nil, err
		}

		addr, err := resolvePingAddr("ip6", icmpNetwork, remoteAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "Trying to resolve %v", remoteAddr)
		}

		return &pingerV6{pingerBase: *newPingerBase(identifier, packetConn, addr, icmpNetwork)}, nil
	}

	return nil, fmt.Errorf("unsupported network type: %v", icmpNetwork)
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

func (p *pingerV4) Ping(seq int, perPingDuration time.Duration) PingResponse {
	return p.ping(seq, ipv4.ICMPTypeEcho, perPingDuration)
}

func (p *pingerV6) Ping(seq int, perPingDuration time.Duration) PingResponse {
	return p.ping(seq, ipv6.ICMPTypeEchoRequest, perPingDuration)
}
