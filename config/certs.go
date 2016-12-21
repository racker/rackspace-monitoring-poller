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

package config

import (
	"crypto/x509"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

// /C=US/ST=Texas/L=San Antonio/O=Rackspace US, Inc/OU=Cloud Monitoring/CN=monitoring-ca.rackspace.com
var productionCAPems = `
-----BEGIN CERTIFICATE-----
MIIEnjCCA4agAwIBAgIJAIkdzmnoN5HMMA0GCSqGSIb3DQEBBQUAMIGQMQswCQYD
VQQGEwJVUzEOMAwGA1UECBMFVGV4YXMxFDASBgNVBAcTC1NhbiBBbnRvbmlvMRow
GAYDVQQKExFSYWNrc3BhY2UgVVMsIEluYzEZMBcGA1UECxMQQ2xvdWQgTW9uaXRv
cmluZzEkMCIGA1UEAxMbbW9uaXRvcmluZy1jYS5yYWNrc3BhY2UuY29tMB4XDTEy
MDYxMzIzMTIzMVoXDTMyMDYxMjIzMTIzMVowgZAxCzAJBgNVBAYTAlVTMQ4wDAYD
VQQIEwVUZXhhczEUMBIGA1UEBxMLU2FuIEFudG9uaW8xGjAYBgNVBAoTEVJhY2tz
cGFjZSBVUywgSW5jMRkwFwYDVQQLExBDbG91ZCBNb25pdG9yaW5nMSQwIgYDVQQD
Exttb25pdG9yaW5nLWNhLnJhY2tzcGFjZS5jb20wggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQDd28woPKcvzJqYGkMF/on9l5sHpPVqf0W2emaL+yatWnbg
I1bZ3NWiyvD8fQTAQY81rwVFyo9Eszv+ct4KKESFalydPgkY5j2E6M/wWHDoEHMe
t7nqYtM4JygaJMm+mqsVMUUCzNHH1oIsPT1bMWlnGH0/P+TC1e01DxwT96wg8nLw
QP3NUKb3Ce/CYwPZtKpTb1jFtXbHkW6TKwtOmVhwnENKJbX8VmMgSCQ74lgWE6eI
J9qA1fH0bSTpPXrKiLBGmdy/ut9WmBG/wmXS0862l1qug9HpvXgZemuAG3g2VarG
9/EUWI/ZyzpYs9QlWW2olIPB2koZKaOuoL0l3AyjAgMBAAGjgfgwgfUwHQYDVR0O
BBYEFMyiM3RcBKbstse0vy3cgz6IuMV+MIHFBgNVHSMEgb0wgbqAFMyiM3RcBKbs
tse0vy3cgz6IuMV+oYGWpIGTMIGQMQswCQYDVQQGEwJVUzEOMAwGA1UECBMFVGV4
YXMxFDASBgNVBAcTC1NhbiBBbnRvbmlvMRowGAYDVQQKExFSYWNrc3BhY2UgVVMs
IEluYzEZMBcGA1UECxMQQ2xvdWQgTW9uaXRvcmluZzEkMCIGA1UEAxMbbW9uaXRv
cmluZy1jYS5yYWNrc3BhY2UuY29tggkAiR3Oaeg3kcwwDAYDVR0TBAUwAwEB/zAN
BgkqhkiG9w0BAQUFAAOCAQEAImybXzvy1UWGmWUNFb98GkVCwGoX6jtbPqwjmB8k
7SFfr7x3nVFKv+2OIbu6w/b00iXMt+ml5s4TOacgzl8KtQ+VQIKoaOA6xhaPwizN
lRPmUWc7Vk3ToR5kBc67zJ4t88FjVro4qtBgTbJRCeKrdc81CVzGeLvEdf5QV8Dm
OG39Ga+BmtT5G3cc0iJXn66Mibiqt9a+ShzPnHYe3A6ChrYuqC79wE17reiwdbSH
zgLLHz7VcQsENuks54XQYegqh/15pZeXGPP4DZax9uTE1vje+rs1JPFXmXqyfmYi
b5004t1fGsVBrTfYAVKb71hPrcstUlKNAokaRYmnqsFYvQ==
-----END CERTIFICATE-----
`

// /C=US/ST=Texas/L=San Antonio/O=Rackspace US, Inc/OU=Cloud Monitoring/CN=monitoring-test-ca.rackspace.com
var staginCAPems = `
-----BEGIN CERTIFICATE-----
MIIErTCCA5WgAwIBAgIJAOvXICx23IUlMA0GCSqGSIb3DQEBBQUAMIGVMQswCQYD
VQQGEwJVUzEOMAwGA1UECBMFVGV4YXMxFDASBgNVBAcTC1NhbiBBbnRvbmlvMRow
GAYDVQQKExFSYWNrc3BhY2UgVVMsIEluYzEZMBcGA1UECxMQQ2xvdWQgTW9uaXRv
cmluZzEpMCcGA1UEAxMgbW9uaXRvcmluZy10ZXN0LWNhLnJhY2tzcGFjZS5jb20w
HhcNMTIwNjEzMjMxNTQwWhcNMzIwNjEyMjMxNTQwWjCBlTELMAkGA1UEBhMCVVMx
DjAMBgNVBAgTBVRleGFzMRQwEgYDVQQHEwtTYW4gQW50b25pbzEaMBgGA1UEChMR
UmFja3NwYWNlIFVTLCBJbmMxGTAXBgNVBAsTEENsb3VkIE1vbml0b3JpbmcxKTAn
BgNVBAMTIG1vbml0b3JpbmctdGVzdC1jYS5yYWNrc3BhY2UuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2qEkUQr7qU+yPih1pqRGApFMfUXevt12
2p4jeY9Nvf9wpaR3TJE70+eZOE95hqRx9xSBxOGooBlG83nj2oKbUhJtOfCvq95y
S7E9GR0gsL03UtNJ0HH6fNk0wuUCQro3xdPZvYuzs7hFPU3qlaY4Yl2PBzgwmo6n
yMLZ4OGX4m4deNKc0bBLe/vn0Gp5hE2yjkwU5fKSbM1GK1QGEBEJ+8iiyS9P+6Xg
1XDAZGzfKBwRTFQWZ+B4W9ZnkbMsnqTP7zocNIUIbu7kolTUHJwaTBYllS+kTO3l
b7KkUBGWGiGCZth8wqSrmBqVM/pHvmfvD7uwy90+jnP6NpFPiYjqaQIDAQABo4H9
MIH6MB0GA1UdDgQWBBRqlKi81uAp3UCEWEU+NZX7qD6fnzCBygYDVR0jBIHCMIG/
gBRqlKi81uAp3UCEWEU+NZX7qD6fn6GBm6SBmDCBlTELMAkGA1UEBhMCVVMxDjAM
BgNVBAgTBVRleGFzMRQwEgYDVQQHEwtTYW4gQW50b25pbzEaMBgGA1UEChMRUmFj
a3NwYWNlIFVTLCBJbmMxGTAXBgNVBAsTEENsb3VkIE1vbml0b3JpbmcxKTAnBgNV
BAMTIG1vbml0b3JpbmctdGVzdC1jYS5yYWNrc3BhY2UuY29tggkA69cgLHbchSUw
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOCAQEAXwhak85QkVmfIc+pCpA6
a7TGCFLmRx8XXOFuiZJ8fK9tVnSzeXIQJJcwuHhkEv6k84O5qyn2qAUtyIih5WhH
8uYC9CtB4jwOP7T9bnnyIwZacyoKctyzXSyS7wGiaxDJLjRCvLwQLzrSg9ighxIl
f6bne3zC74iOGeCEHjcMh6dHfAY4GBfDTL8dNq0xykzVhOViz9PQIoG9sWYdx5y3
wS0chXM06TC4sQmFIpQL8/rbbiQLe5946msQm88m+++MSLiieWagduSUyZZpvYBy
hIqt44pK2q73HXC7rRMXYqRBl7HcToJX7ZuG6Dub0cBrtUhCn4tahShKqt2oQ6yE
FA==
-----END CERTIFICATE-----
`

func loadFromString(certPems string) *x509.CertPool {
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(certPems))
	if !ok {
		return nil
	}
	return certPool
}

func LoadProductionCAs() *x509.CertPool {
	return loadFromString(productionCAPems)
}

func LoadStagingCAs() *x509.CertPool {
	return loadFromString(staginCAPems)
}

func LoadDevelopmentCAs(pemFilename string) *x509.CertPool {
	content, err := ioutil.ReadFile(pemFilename)
	if err != nil {
		logrus.WithField("pemFilename", pemFilename).Error("Unable to load development cert pool")
		return nil
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(content)
	if !ok {
		return nil
	}
	return certPool

}
