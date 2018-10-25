/*
Copyright (C) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tls

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/docker/machine/libmachine/drivers"
)

var (
	CACert = []byte(`-----BEGIN CERTIFICATE-----
MIIF7jCCA9agAwIBAgIJAJ6zwIei4Z1/MA0GCSqGSIb3DQEBCwUAMIGLMQswCQYD
VQQGEwJVUzESMBAGA1UECAwJTWluaXNoaWZ0MRIwEAYDVQQKDAlNaW5pc2hpZnQx
GzAZBgNVBAsMEkludGVybWVkaWF0ZSBwcm94eTEVMBMGA1UEAwwMbWluaXNoaWZ0
LmlvMSAwHgYJKoZIhvcNAQkBFhFpbmZvQG1pbmlzaGlmdC5pbzAeFw0xODA4MzAx
OTM4MjhaFw0zODA4MjUxOTM4MjhaMIGLMQswCQYDVQQGEwJVUzESMBAGA1UECAwJ
TWluaXNoaWZ0MRIwEAYDVQQKDAlNaW5pc2hpZnQxGzAZBgNVBAsMEkludGVybWVk
aWF0ZSBwcm94eTEVMBMGA1UEAwwMbWluaXNoaWZ0LmlvMSAwHgYJKoZIhvcNAQkB
FhFpbmZvQG1pbmlzaGlmdC5pbzCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBAMyTDo4trmexv5Py9aTiCpiIsCHUH0PVqKbzfCY5o5b8iMcBUdhpUIUvh1dz
lQjdY1fJAE9IqyjihQ9OhXB6SiJFai+sEQb1F7F0vEZ3cNCQf3SW74SzUjQTyuiP
waErP9ENOGwX9XgRikXlptBfwyGATMf1joZRCTBqTZVmowNNddu2xJik44CcLfHK
PKcBjTEuVGOBFoKWr/HX6rLpqAgkxoXZosdXCIVqFg3y9XHI2L5bOlxjuGMTz8QR
mg7cZoqxOk8q9bCwBVM52CwrgRFrEv/9UhAzosK8MRSLSbF95px/Ihft0g0XtGXz
pFMJovDGwbTfj3xyBKmcsdqGP5Z9Ro6sblb/nX+HHA8vCH0+v5Pq1kxcx2mSWyh9
xJQ9ElTu/RLBsjtCmovm4piqm63BhySGSKaZYFcx54Olr4d2+dpFMHf9pWDKF83n
S2zg5egShWkr/WapkrlF5KMOQC2GiFd3jcED6wJ1HewGVSn5jXzK7yhJ0kcI90iW
tL6vc5N3fLqYMQrTIOWUJewshuNh9SDt9DrVOgl3xJPm+dlrWuHW1g3ru/jv6syE
Dcovqb6cxIOZCXoaHqxABya4bnNb3m4sKQ8iN51OAtP7YJz0YtmbL8b0GZ47b8Li
jP9jSc55NqusWRpZeckLFMZeQvNLN0p8doN37ONZNeKkdhLVAgMBAAGjUzBRMB0G
A1UdDgQWBBSTpHZvXeaSpQB3WI/alesnLcBrJjAfBgNVHSMEGDAWgBSTpHZvXeaS
pQB3WI/alesnLcBrJjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IC
AQBfF7JCZC/q4jNF2+grIWvhplkA/lWgy4dCt4+UoGvRtMbnNJICTVv2nFaoDe9E
6DVQusBmKhw2RmGBeiigIC30f3tFqijbLoY17J0/htIODNnjY6MAJWHI8PyhTkbO
1efGsKncoHAvLV5qnGmb7pOSGBTUlhF5MoBMMaB5d888S0IN9z3QkoAV4LR1TeMT
fEYILlr2xOE20+V/CWZ8DlQ1IRzYXi1ZSd5P0h90IVy4LX5EzX201VGJWaoS6p7U
3TTwgm8ZtjnqI4flSB6jV2OwQXs7ghbmzG9BcxNbW7s4MAeVznwlGgfLLsCFJ8gI
8o3tIY+sRmHsPF/th1vmKdLMULZ8muX9sVD6N9uNi4NAKJ+WYrVMuF/YcqwL6KPW
brKHduvWgMCdcGcm0eMDMAVS5NAsKhZBTWH6R1V66OCFDM95myUrgc/ACd904g5d
+BlDDAOMKadJxI5pWIO9zzz4zNyw0kInzVYZ6SCKL8dJehioRWHAFBRI1RNRIUGM
9wmYY9jq8BrAjFmJcUdTW/7AOGvaNr3QXlmaU+WjMT5of802G3eYAFeCoSp4AOsl
WmAliKC2nKWYhpaESuNPpJi/wg2fD9uRN8cpsuc0yMZFhYZSlY4a0aObXawRnKVy
n3c/ezDP9vhy5Dt0PZLW8RRj5oAXbwRRdiVVfGNRcFOQew==
-----END CERTIFICATE-----`)
	CAKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`)
)

func SetCACertificate(driver drivers.Driver, certificate []byte) bool {
	encodedCertificate := base64.StdEncoding.EncodeToString(certificate)

	cmd := fmt.Sprintf(
		"echo %s | base64 --decode | sudo tee -a /etc/pki/tls/certs/ca-bundle.crt > /dev/null",
		encodedCertificate)

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		return false
	}

	return true
}

func SetCACertificateFromWatcher(driver drivers.Driver, certificate []byte) bool {
	encodedCertificate := base64.StdEncoding.EncodeToString(certificate)

	// for some reason filtering does not work here
	watchCmd := fmt.Sprintf("dockerwatch exec -- bash -c 'echo %s | base64 --decode | tee -a /etc/pki/tls/certs/ca-bundle.crt' &", encodedCertificate)
	go drivers.RunSSHCommandFromDriver(driver, watchCmd) // to allow the command happen in the background (need callback to ensure watchdog behaviour)

	return true
}

func LoadCertificate(filepath string) ([]byte, error) {

	output, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return output, nil
}
