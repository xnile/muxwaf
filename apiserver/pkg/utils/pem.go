package utils

import "regexp"

var pemCertificatePattern = regexp.MustCompile(`^(-{5}BEGIN CERTIFICATE-{5}\x{000D}?\x{000A}([A-Za-z0-9/+]{64}\x{000D}?\x{000A})*[A-Za-z0-9/+]{1,64}={0,2}\x{000D}?\x{000A}-{5}END CERTIFICATE-{5}\x{000D}?\x{000A})*-{5}BEGIN CERTIFICATE-{5}\x{000D}?\x{000A}([A-Za-z0-9/+]{64}\x{000D}?\x{000A})*[A-Za-z0-9/+]{1,64}={0,2}\x{000D}?\x{000A}-{5}END CERTIFICATE-{5}(\x{000D}?\x{000A})?$`)
var pemPrivateKeyPattern = regexp.MustCompile(`^-{5}BEGIN (RSA|EC) PRIVATE KEY-{5}\x{000D}?\x{000A}([A-Za-z0-9/+]{64}\x{000D}?\x{000A})*[A-Za-z0-9/+]{1,64}={0,2}\x{000D}?\x{000A}-{5}END (RSA|EC) PRIVATE KEY-{5}(\x{000D}?\x{000A})?$`)

func IsValidPEMCertificate(cert string) bool {
	return pemCertificatePattern.MatchString(cert)
}

func IsValidPEMPrivateKey(key string) bool {
	return pemPrivateKeyPattern.MatchString(key)
}
