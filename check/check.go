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

	"io"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	protocheck "github.com/racker/rackspace-monitoring-poller/protocol/check"
	"github.com/racker/rackspace-monitoring-poller/protocol/metric"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

// Check is an interface required to be implemented by all checks.
// Implementations of this interface should extend CheckBase, which leaves only the Run method to be implemented.
type Check interface {
	GetId() string
	SetId(id string)
	GetEntityId() string
	GetCheckType() string
	SetCheckType(checkType string)
	GetPeriod() uint64
	GetWaitPeriod() time.Duration
	SetPeriod(period uint64)
	GetTimeout() uint64
	GetTimeoutDuration() time.Duration
	SetTimeout(timeout uint64)
	Cancel()
	Done() <-chan struct{}
	Run() (*CheckResultSet, error)
}

var WaitPeriodTimeMeasurement = time.Second

// CheckBase provides an abstract implementation of the Check interface leaving Run to be implemented.
type CheckBase struct {
	protocheck.CheckIn
	// context is primarily provided to enable watching for cancellation of this particular check
	context context.Context
	// cancel is associated with the context and can be invoked to initiate the cancellation
	cancel context.CancelFunc
}

// GetTargetIP obtains the specific IP address selected for this check.
// It returns the resolved IP address as dotted string.
func (ch *CheckBase) GetTargetIP() (string, error) {
	ip, ok := ch.IpAddresses[*ch.TargetAlias]
	if ok {
		return ip, nil
	}
	return "", errors.New("Invalid Target IP")

}

func (ch *CheckBase) GetId() string {
	return ch.Id
}

func (ch *CheckBase) SetId(id string) {
	ch.Id = id
}

func (ch *CheckBase) GetCheckType() string {
	return ch.CheckType
}

func (ch *CheckBase) SetCheckType(checkType string) {
	ch.CheckType = checkType
}

func (ch *CheckBase) GetPeriod() uint64 {
	return ch.Period
}

func (ch *CheckBase) GetEntityId() string {
	return ch.EntityId
}

func (ch *CheckBase) SetPeriod(period uint64) {
	ch.Period = period
}

func (ch *CheckBase) GetTimeout() uint64 {
	return ch.Timeout
}

func (ch *CheckBase) GetTimeoutDuration() time.Duration {
	return time.Duration(ch.Timeout) * time.Second
}

func (ch *CheckBase) SetTimeout(timeout uint64) {

}

func (ch *CheckBase) GetWaitPeriod() time.Duration {
	return time.Duration(ch.Period) * WaitPeriodTimeMeasurement
}

func (ch *CheckBase) Cancel() {
	ch.cancel()
}

func (ch *CheckBase) Done() <-chan struct{} {
	return ch.context.Done()
}

type TLSMetrics struct {
	Verified bool
}

func (ch *CheckBase) AddTLSMetrics(cr *CheckResult, state tls.ConnectionState) (*TLSMetrics, error) {
	tlsMetrics := &TLSMetrics{}
	if len(state.PeerCertificates) == 0 {
		return tlsMetrics, nil
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
	// SERIAL
	cr.AddMetric(metric.NewMetric("cert_serial", "", metric.MetricNumber, cert.SerialNumber, ""))
	if len(cert.OCSPServer) > 0 {
		cr.AddMetric(metric.NewMetric("cert_ocsp", "", metric.MetricNumber, cert.OCSPServer[0], ""))
	}
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		publicKey := cert.PublicKey.(*rsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.N.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "rsa", ""))
	case x509.DSA:
		publicKey := cert.PublicKey.(*dsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Q.BitLen(), ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "dsa", ""))
	case x509.ECDSA:
		publicKey := cert.PublicKey.(*ecdsa.PublicKey)
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, publicKey.Params().BitSize, ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "ecdsa", ""))
	default:
		cr.AddMetric(metric.NewMetric("cert_bits", "", metric.MetricNumber, "0", ""))
		cr.AddMetric(metric.NewMetric("cert_type", "", metric.MetricNumber, "-", ""))
	}
	// CERT SIG ALGO
	cr.AddMetric(metric.NewMetric("cert_sig_algo", "", metric.MetricNumber, strings.ToLower(cert.SignatureAlgorithm.String()), ""))
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
	cr.AddMetric(metric.NewMetric("ssl_session_version", "", metric.MetricNumber, sslVersion, ""))
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
	cr.AddMetric(metric.NewMetric("ssl_session_cipher", "", metric.MetricNumber, cipherSuite, ""))
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
	return tlsMetrics, nil
}

func (ch *CheckBase) readLimit(conn io.Reader, limit int64) ([]byte, error) {
	bytes := make([]byte, limit)
	bio := io.LimitReader(conn, limit)
	count, err := bio.Read(bytes)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return bytes[:count], nil
}

func (ch *CheckBase) PrintDefaults() {
	var targetAlias string
	var targetHostname string
	var targetResolver string
	if ch.TargetAlias != nil {
		targetAlias = *ch.TargetAlias
	}
	if ch.TargetHostname != nil {
		targetHostname = *ch.TargetHostname
	}
	if ch.TargetResolver != nil {
		targetResolver = *ch.TargetResolver
	}
	log.WithFields(log.Fields{
		"type":            ch.CheckType,
		"period":          ch.Period,
		"timeout":         ch.Timeout,
		"disabled":        ch.Disabled,
		"ipaddresses":     ch.IpAddresses,
		"target_alias":    targetAlias,
		"target_hostname": targetHostname,
		"target_resolver": targetResolver,
		"details":         string(*ch.RawDetails),
	}).Infof("New check %v", ch.GetId())
}
