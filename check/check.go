//
// Copyright 2016 Rackspace
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

// Package check contains an implementation file per check-type.
//
// A typical/starter check implementation file would contain something like:
//
//  type <TYPE>Check struct {
//    CheckBase
//    Details struct {
//      SomeField string|... `json:"some_field"`
//      ...
//    }
//  }
//
//  func New<TYPE>Check(base *CheckBase) Check {
//    check := &<TYPE>Check{CheckBase: *base}
//    err := json.Unmarshal(*base.Details, &check.Details)
//    if err != nil {
//      log.Printf("Error unmarshalling check details")
//      return nil
//    }
//    check.PrintDefaults()
//    return check
//  }
//
//  func (ch *<TYPE>Check) Run() (*CheckResultSet, error) {
//    log.Printf("Running <TYPE> Check: %v", ch.GetId())
//    ...do check specifics...
//
//    ...upon success:
//    cr := NewCheckResult(
//      metric.NewMetric(...)
//      ...
//    )
//    crs := NewCheckResultSet(ch, cr)
//    crs.SetStateAvailable()
//    return crs, nil
//  }
package check

import (
	"context"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"

	"fmt"
	"io"
	"net"
	"strings"

	"time"

	"encoding/json"
	protocheck "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"io/ioutil"
)

// Check is an interface required to be implemented by all checks.
// Implementations of this interface should extend CheckBase,
// which leaves only the Run method to be implemented.
// Run method needs to set up the check,

// send the request, and parse the response
type Check interface {
	fmt.Stringer

	utils.LogPrefixGetter
	GetID() string
	SetID(id string)
	GetEntityID() string
	GetZoneID() string
	GetCheckType() string
	GetTargetIP() (string, error)
	SetCheckType(checkType string)
	GetPeriod() uint64
	GetWaitPeriod() time.Duration
	SetPeriod(period uint64)
	GetTimeout() uint64
	GetTimeoutDuration() time.Duration
	SetTimeout(timeout uint64)
	IsDisabled() bool
	Cancel()
	Done() <-chan struct{}
	Run() (*ResultSet, error)
	GetCheckIn() *protocheck.CheckIn
}

const (
	ResolverIPV4 = 1
	ResolverIPV6 = 2
)

// The prototype for a dial context
type DialContextFunc func(context.Context, string, string) (net.Conn, error)

// WaitPeriodTimeMeasurement sets up the time measurement
// for how long the check should wait before retrying.
// By default, it'll wait for check's "period" seconds
var WaitPeriodTimeMeasurement = time.Second

// ErrInvalidTargetIP invalid target IP
var ErrInvalidTargetIP = errors.New("Invalid Target IP")

// Base provides an abstract implementation of the Check
// interface leaving Run to be implemented.
type Base struct {
	protocheck.CheckIn
	// context is primarily provided to enable watching for cancellation of this particular check
	context context.Context
	// cancel is associated with the context and can be invoked to initiate the cancellation
	cancel context.CancelFunc
}

// GetTargetIP obtains the specific IP address selected for this check.
// It returns the resolved IP address as dotted string.
func (ch *Base) GetTargetIP() (string, error) {
	if ch.TargetAlias != nil {
		ip, ok := ch.IpAddresses[*ch.TargetAlias]
		if ok {
			return ip, nil
		}
	} else if ch.TargetHostname != nil && *ch.TargetHostname != "" {
		return *ch.TargetHostname, nil
	}
	return "", ErrInvalidTargetIP
}

// GetID  returns check's id
func (ch *Base) GetID() string {
	return ch.Id
}

// SetID sets check's id to provided id
func (ch *Base) SetID(id string) {
	ch.Id = id
}

// GetCheckType returns check's type
func (ch *Base) GetCheckType() string {
	return ch.CheckType
}

// GetLogPrefix returns the log prefix
func (ch *Base) GetLogPrefix() string {
	return fmt.Sprintf("%v:%v", ch.GetID(), ch.GetCheckType())
}

func (ch *Base) String() string {
	return fmt.Sprintf("[id:%v, type:%v]", ch.GetID(), ch.GetCheckType())
}

// SetCheckType sets check's checktype to
// to provided check type
func (ch *Base) SetCheckType(checkType string) {
	ch.CheckType = checkType
}

// GetPeriod returns check's period
func (ch *Base) GetPeriod() uint64 {
	return ch.Period
}

// SetPeriod sets check's period to
// to provided period
func (ch *Base) SetPeriod(period uint64) {
	ch.Period = period
}

// GetEntityID  returns check's entity id
func (ch *Base) GetEntityID() string {
	return ch.EntityId
}

// GetZoneID returns check's zone ID
func (ch *Base) GetZoneID() string {
	return ch.ZoneId
}

// GetTimeout returns check's timeout
func (ch *Base) GetTimeout() uint64 {
	return ch.Timeout
}

// SetTimeout is not currently implemented
func (ch *Base) SetTimeout(timeout uint64) {

}

// GetTimeoutDuration returns check's timeout in seconds
func (ch *Base) GetTimeoutDuration() time.Duration {
	return time.Duration(ch.Timeout) * time.Second
}

// GetWaitPeriod returns check's period in
// provided time measurements.  Defaulted to seconds
func (ch *Base) GetWaitPeriod() time.Duration {
	return time.Duration(ch.Period) * WaitPeriodTimeMeasurement
}

func (ch *Base) IsDisabled() bool {
	return ch.Disabled
}

// Cancel method closes the channel's context
func (ch *Base) Cancel() {
	ch.cancel()
}

// Done method listens on channel's context to close
func (ch *Base) Done() <-chan struct{} {
	return ch.context.Done()
}

// GetCheckIn resolves the underlying instance
func (ch *Base) GetCheckIn() *protocheck.CheckIn {
	return &ch.CheckIn
}

// TLSMetrics is utilized the provide TLS metrics
type TLSMetrics struct {
	Verified bool
}

// AddTLSMetrics function sets up secure metrics from provided check result
// It then returns a list of tls metrics with optional Verified parameter
// that depends on whether the certificate was successfully signed (based on provided tls
// connection state)
func (ch *Base) AddTLSMetrics(cr *Result, state tls.ConnectionState) *TLSMetrics {
	// TODO: refactor.  Cyclomatic complexity of 35
	tlsMetrics := &TLSMetrics{}
	if len(state.PeerCertificates) == 0 {
		return tlsMetrics
	}
	// Validate certificate chain
	cert := state.PeerCertificates[0]
	opts := x509.VerifyOptions{
		Roots:         nil,
		CurrentTime:   time.Now(),
		DNSName:       state.ServerName,
		Intermediates: x509.NewCertPool(),
	}
	for i, cert := range state.PeerCertificates {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}
	_, err := cert.Verify(opts)
	if err != nil {
		tlsMetrics.Verified = false
		cr.AddMetric(metric.NewMetric("cert_error", "", metric.MetricString, err.Error(), ""))
	} else {
		tlsMetrics.Verified = true
	}
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		publicKey := cert.PublicKey.(*rsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.N.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricString, "rsa", ""))
	case x509.DSA:
		publicKey := cert.PublicKey.(*dsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Q.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricString, "dsa", ""))
	case x509.ECDSA:
		publicKey := cert.PublicKey.(*ecdsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Params().BitSize, ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricString, "ecdsa", ""))
	default:
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, "0", ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricString, "-", ""))
	}
	// CERT SIG ALGO
	cr.AddMetric(metric.NewMetric("cert_sig_algo", "", metric.MetricString, strings.ToLower(cert.SignatureAlgorithm.String()), ""))
	var sslVersion string
	switch state.Version {
	case tls.VersionSSL30:
		sslVersion = "ssl3"
	case tls.VersionTLS10:
		sslVersion = "tls1.0"
	case tls.VersionTLS11:
		sslVersion = "tls1.1"
	case tls.VersionTLS12:
		sslVersion = "tls1.2"
	default:
		sslVersion = "-"
	}
	// SESSION VERSION
	cr.AddMetric(metric.NewMetric("ssl_session_version", "", metric.MetricString, sslVersion, ""))
	var cipherSuite string
	switch state.CipherSuite {
	case tls.TLS_RSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_RSA_WITH_RC4_128_SHA"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_RC4_128_SHA"
	case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"

	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		cipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	default:
		cipherSuite = "-"
	}
	// SESSION CIPHER
	cr.AddMetric(metric.NewMetric("ssl_session_cipher", "", metric.MetricString, cipherSuite, ""))
	// ISSUER
	if issuer, err := utils.GetDNFromCert(cert.Issuer, "/"); err == nil {
		cr.AddMetric(metric.NewMetric("cert_issuer", "", metric.MetricString, issuer, ""))
	}
	// SUBJECT
	if subject, err := utils.GetDNFromCert(cert.Subject, "/"); err == nil {
		cr.AddMetric(metric.NewMetric("cert_subject", "", metric.MetricString, subject, ""))
	}
	// ALTERNATE NAMES
	cr.AddMetric(metric.NewMetric("cert_subject_alternate_names", "", metric.MetricString, strings.Join(cert.DNSNames, ", "), ""))
	// START TIME
	cr.AddMetric(metric.NewMetric("cert_start", "", metric.MetricNumber, cert.NotBefore.Unix(), ""))
	cr.AddMetric(metric.NewMetric("cert_end", "", metric.MetricNumber, cert.NotAfter.Unix(), ""))
	return tlsMetrics
}

func (ch *Base) readLimit(conn io.Reader, limit int64) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(conn, limit+1))
	if err != nil && err != io.EOF {
		return nil, err
	}
	return body, nil
}

func ReadCheckFromFile(filename string) (Check, error) {
	jsonContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return NewCheck(ctx, json.RawMessage(jsonContent))
}

func NewCustomDialContext(rType uint64) DialContextFunc {
	netType := "tcp4"
	switch rType {
	case ResolverIPV4:
		netType = "tcp4"
	case ResolverIPV6:
		netType = "tcp6"
	default:
	}
	return func(ctx context.Context, network string, address string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, netType, address)
	}
}
