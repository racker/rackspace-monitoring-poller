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
package utils

import (
	"crypto/x509/pkix"
	"fmt"
	"strings"
)

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)

var IntegrationTestCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDBDCCAeygAwIBAgIJAJ/tbXP5LIrVMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwMTEyMTYyNzU3WhcNMjIwMTExMTYyNzU3WjAU
MRIwEAYDVQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDmG/LeO6b6dO5iftb2vWwolX4CM3vUvTZZr1ocq67yUWeH5xdCUwIcoBwf
yzi8DDv2xPLoftCXk5WhmpwzbAeDx4UiN5RFaHsF58pkLj0dgisUDf9lcT4t5qWe
ZZL2uas+N7GoToCSpvoIdujTCgwx395s8fsySclPudpk10kx6UCguoAn4/WHSlFE
DURRpMzews9k3D2lFdpA7VyUaMrI6IioUOs4rOK65g68xlr5uW685ODQWPkz1oQX
OsojMgC+zWb7T08SL7zDiQu2sWr8upWbBfi5epqWY+PqqxVUuQ1Uowcfgi91xq8O
R86QdFWPcHgWRQ2mn05SlAsFV/hTAgMBAAGjKDAmMA8GA1UdEQQIMAaHBH8AAAEw
EwYDVR0lBAwwCgYIKwYBBQUHAwEwDQYJKoZIhvcNAQELBQADggEBABiNlbjqJV86
fJO6WDCQuUXypeEdWxE2bB/IoRTsz4Br9N56Mhk0isdDhd7Iix7nmhiR9thIXzng
uAO+WUmbyTfT7O5aJU62bVRqU+k+OhjVD8DlE7+rftr7w7+LIFUKfO9nb/ogbaZa
wE7tUA+hBvuOFx/GK/bz9WOBTNmfGqx1RoioDOSZZzJag2IoKgu/k84b5jQ3MwAX
SRaoY+e0wptuMPl2te/K9U+0Pe767fll+Vq/NzJxWVCvuyoTGu8AKcO1wZPVIu9d
Pps9FPwZB2Qio9XrjuQyn/PfYrKId+wl9MEC8n+PxVhP8DskWZI5PW9E3uyMn8Nu
TbbKwFgzDuE=
-----END CERTIFICATE-----`)

var IntegrationTestKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA5hvy3jum+nTuYn7W9r1sKJV+AjN71L02Wa9aHKuu8lFnh+cX
QlMCHKAcH8s4vAw79sTy6H7Ql5OVoZqcM2wHg8eFIjeURWh7BefKZC49HYIrFA3/
ZXE+LealnmWS9rmrPjexqE6Akqb6CHbo0woMMd/ebPH7MknJT7naZNdJMelAoLqA
J+P1h0pRRA1EUaTM3sLPZNw9pRXaQO1clGjKyOiIqFDrOKziuuYOvMZa+bluvOTg
0Fj5M9aEFzrKIzIAvs1m+09PEi+8w4kLtrFq/LqVmwX4uXqalmPj6qsVVLkNVKMH
H4IvdcavDkfOkHRVj3B4FkUNpp9OUpQLBVf4UwIDAQABAoIBACzlmR55PxwxAm4f
V2vvC5JjkKF3UBrzDA61ovxjFxBah7vBgA1Fyuyw5KvjZ99w96YvSUHJtINOnWxZ
kU6LLnAs1rIVbA2a1B4T2q5vQydlxWf1TzaIwNwN25SrNuCC24GZNkWjg3yZrcFH
CihbFoQIrQpOsHdgZDH1DkKMqtBc1rEV9zMK5KuzVMOzdzqskLcY2zKSIs2Lmhw/
Vd8Fm+IZCMvLdM1VlDsJRvejQdpK4Rtn79beX+QAM5RtuFFuLqHIy3sfUlfjwLCw
T+NDjjlNMqXqCADnhp7584i/tXrmkLTBmoWBn1kLOrOeQFHiL03/g0N6ehi4Ey0Y
od/7aYkCgYEA9nogywgclqG1yuXSLJ5GHmBlksjd67uFUOlgB3S/IILAwqf7lD8c
h6kp2kOOXwGhFR1BP7FhJvj3+4UgJz0IAbL0aj37ZCtUiomXbyi94AIQmXe4Iyuj
VEB/4odsVQrrnZ9SXfoS6vt1gsBd5LgDiRywSWoBIDTk6ZiMnZTdgrcCgYEA7v/t
kSW+NQW7mGLANYxVlae2+2UCzArzTb6WerwXjwMhAjnTotYKScynpN6On1gT9Sfl
jSoilub2U3bUkfqYRFt63covT6POZD+RyF4S560Yx0gwGwvCGkz9eyhHxpUeOe1d
pfsWuHE3XzgXrmHj3ynhXuwOKVI2cQknOGnJK0UCgYEApuOkxrS0Xs4aAMtCV1HH
2pOc0xnNIeuz5khO7F2BeGrwSB1j/EoLcFP7cb1ibjP1NQ28+3qIdNIJXzYRwl/R
xwy78CAN0xJ/yNpHPk4Q2terE677cF0A13Bg5yqZELA3P1/8boOAQbmIJMNKEC8E
vdc+CkeLgZovEXhoZd7BadsCgYEAkBPQz3OFWsl98bt2S9GxtmpIsPyP1xmy2udO
J+dD/H7SY1kg8EVAJoUtewJ/0Cd0wJGwnI0OFRJe5Kn6M5ZyPKM5SoMcSlJhlaWM
6NFtbCS5j0lBVsyb0uce2CPMQTab5ifmEK1xYPc/fjN+cy2oBVxl9KcxUk+xaisu
bZ+4GlECgYBgmowKyrtG7LOoZ3PB8Y+xzvEg54vIxx0tb+xHY+TsDwFykQ64OC0W
KYHCMCMEpEOlwGcOm6mTh1uE8TRxn8En4zWlsVo1DSdOiCya7LUp1Vhp4ED2eZlc
kUw0dOqL+1istTBjzQGIjZ6W1+VGhTGbfPubms6Bxml2zIKs924ymA==
-----END RSA PRIVATE KEY-----`)

var IntegrationTestCA = []byte(`-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAIMtTD12XoEmMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwMTEyMTYyNzQyWhcNMjIwMTExMTYyNzQyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEArCoS/pJSCWRaqZOTDgbFcup5/EIHe4gGbK9i6FpESWXpcl9Tebb+gkDV
uDfBdrjZhIrLx+jGqhaMLedAS0xGYeWt0IZPYgQ1jw5Jv34mapmCVGclvkNNPY0y
KUxl8O11yy5rOPwmlI7hMQXzgy26laOrzviqyetYs3CaFUrRHQmb6LoexnnylxdS
sArRzm4mHJA6crWf7cqd7iLiqZ9gDBLjBuBXeAOPHXCOKW6aqP9JqiFge/3GaiKt
QHz+Ude1wztGlliWlAb1CI6h50vc1pC7rJbQ3MpOYIxVSTn0oK8bvUodi5h0Rpl8
1KI69w8ic+JIb47iCTZZ+4Zg8SHhfwIDAQABo1AwTjAdBgNVHQ4EFgQU5SItQINr
wsmguiAXUu/Y/rWXO8MwHwYDVR0jBBgwFoAU5SItQINrwsmguiAXUu/Y/rWXO8Mw
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAhCqq17WZmksFX70k1trI
WFKPaITMZPhuTZeF2tRLz31NKQ8U4umi5AQ881FzSglYpBKTuKOB00vCvXKdfxaw
IFcFe8ZkGIyjGYkJPyRZSraLLR3W7l7/Buu6/LUeBKa//iEa4LAkdMmtys6hyUBN
vAwMzQF+92A1U8w+ZLQxhQWcTaWIJHxghTkv+kJIP6g+r0IImt2t1+Vo1q7xmlB2
azbvhVrag1fAhVrE7YkYtGbXycZgybUjTBuHUDtRxmfZSxTFtoBBQF+KWxrWOs+N
WAVnloOYRpyE+FSK0T3ySbf92HSyZtT0zC9I1+X0VUzbFdZYUlu5WNDpHREFvpXK
vQ==
-----END CERTIFICATE-----`)

var oid = map[string]string{
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "streetAddress",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.17":                   "postalCode",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"0.9.2342.19200300.100.1.1":  "userid",
}

// Convert a DN to a String
// return the string, or and error
func GetDNFromCert(namespace pkix.Name, sep string) (string, error) {
	subject := []string{}
	for _, s := range namespace.ToRDNSequence() {
		for _, i := range s {
			if v, ok := i.Value.(string); ok {
				if name, ok := oid[i.Type.String()]; ok {
					// <oid name>=<value>
					subject = append(subject, fmt.Sprintf("%s=%s", name, v))
				} else {
					// <oid>=<value> if no <oid name> is found
					subject = append(subject, fmt.Sprintf("%s=%s", i.Type.String(), v))
				}
			} else {
				// <oid>=<value in default format> if value is not string
				subject = append(subject, fmt.Sprintf("%s=%v", i.Type.String, v))
			}
		}
	}
	return sep + strings.Join(subject, sep), nil
}
