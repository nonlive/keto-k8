package etcd

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	certutil "github.com/UKHomeOffice/kmm/pkg/client-go/util/cert"

	"github.com/UKHomeOffice/kmm/pkg/kubeadm/pkiutil"
	"github.com/UKHomeOffice/kmm/pkg/fileutil"
)

type ServerConfig struct {
	CaKeyFileName		string
	ServerCertFileName	string
	ServerKeyFileName	string
	PeerCertFileName	string
	PeerKeyFileName		string
	LocalHostNames		[]string
	ClusterHostNames	[]string
	ClientConfig		ClientConfig
}

// ExtKeyUsage contains a mapping of string names to extended key
// usages.
var ExtKeyUsage = map[string]x509.ExtKeyUsage{
	"any":              x509.ExtKeyUsageAny,
	"server auth":      x509.ExtKeyUsageServerAuth,
	"client auth":      x509.ExtKeyUsageClientAuth,
	"code signing":     x509.ExtKeyUsageCodeSigning,
	"email protection": x509.ExtKeyUsageEmailProtection,
	"s/mime":           x509.ExtKeyUsageEmailProtection,
	"ipsec end system": x509.ExtKeyUsageIPSECEndSystem,
	"ipsec tunnel":     x509.ExtKeyUsageIPSECTunnel,
	"ipsec user":       x509.ExtKeyUsageIPSECUser,
	"timestamping":     x509.ExtKeyUsageTimeStamping,
	"ocsp signing":     x509.ExtKeyUsageOCSPSigning,
	"microsoft sgc":    x509.ExtKeyUsageMicrosoftServerGatedCrypto,
	"netscape sgc":     x509.ExtKeyUsageNetscapeServerGatedCrypto,
}

/*
Missing fields when generating client certs:
            X509v3 Subject Key Identifier:
                BE:AE:DD:86:6F:79:48:9A:33:CE:4D:C0:07:D3:5E:FF:70:02:BB:78
 */

func GenCerts(cfg ServerConfig) (err error) {

	var caCert *x509.Certificate
	var caKey *rsa.PrivateKey

	// Load the CA files...
	if fileutil.ExistFile(cfg.CaKeyFileName) && fileutil.ExistFile(cfg.ClientConfig.CaFileName) {

		// Try to load cert and key...
		caCert, err = pkiutil.TryLoadAnyCertFromDisk(cfg.ClientConfig.CaFileName)
		if err != nil || caCert == nil {
			return fmt.Errorf("CA certificate existed but could not be loaded properly %q", cfg.ClientConfig.CaFileName)
		}
		// The certificate and key could be loaded, but the certificate is not a CA
		if !caCert.IsCA {
			return fmt.Errorf("certificate and key could be loaded but the certificate is not a CA")
		}

		caKey, err = pkiutil.TryLoadAnyKeyFromDisk(cfg.CaKeyFileName)
		if err != nil || caKey == nil {
			return fmt.Errorf("CA key existed but could not be loaded properly %q", cfg.CaKeyFileName)
		}
		fmt.Printf("Found and verified CA certificate %q and key %q.\n", cfg.ClientConfig.CaFileName, cfg.CaKeyFileName)
	}

	// Generate the ETCD server cert and key file (if required)
	serverCertCfg := certutil.Config{
		CommonName: cfg.ClusterHostNames[0],
		AltNames:   getAltNames(cfg.ClusterHostNames),
		Usages:     []x509.ExtKeyUsage{
			ExtKeyUsage["server auth"],
		},
	}
	if err = checkOrCreateCert(
		cfg.ServerCertFileName,
		cfg.ServerKeyFileName,
		caCert,
		caKey,
		serverCertCfg); err != nil {

		return err
	}

	// Generate ETCD peer cert and key (if required)
	peerCertCfg := certutil.Config{
		CommonName: cfg.ClusterHostNames[0],
		AltNames:   getAltNames(cfg.LocalHostNames),
		Usages:     []x509.ExtKeyUsage{
			ExtKeyUsage["server auth"],
			ExtKeyUsage["client auth"],
		},
	}
	if err = checkOrCreateCert(
		cfg.PeerCertFileName,
		cfg.PeerKeyFileName,
		caCert,
		caKey,
		peerCertCfg); err != nil {

		return err
	}

	// Generate ETCD client certs
	clientCertCfg := certutil.Config{
		CommonName: cfg.ClusterHostNames[0],
		Usages:     []x509.ExtKeyUsage{
			ExtKeyUsage["client auth"],
		},
	}
	if err = checkOrCreateCert(
		cfg.ClientConfig.ClientCertFileName,
		cfg.ClientConfig.ClientKeyFileName,
		caCert,
		caKey,
		clientCertCfg); err != nil {
		return err
	}

	return err
}

func checkOrCreateCert(certFile, keyFile string, caCert *x509.Certificate, caKey *rsa.PrivateKey, config certutil.Config) (error) {
	if fileutil.ExistFile(certFile) && fileutil.ExistFile(keyFile) {
		// Try to load cert and key...
		cert, err := pkiutil.TryLoadAnyCertFromDisk(certFile)
		if err != nil || cert == nil {
			return fmt.Errorf("certificate existed but they could not be loaded properly %q", certFile)
		}
		key, err := pkiutil.TryLoadAnyKeyFromDisk(keyFile)
		if err != nil || key == nil {
			return fmt.Errorf("key existed but they could not be loaded properly %q", keyFile)
		}

		fmt.Printf("Using cert:%q and key %q\n", certFile, keyFile)
	} else {
		// The certificate and / or the key did NOT exist, let's generate them now
		cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, config)
		if err != nil {
			return fmt.Errorf("failure while creating key %q and cert %q [%v]", certFile, keyFile, err)
		}
		if err = certutil.WriteCert(certFile, certutil.EncodeCertPEM(cert)); err != nil {
			return fmt.Errorf("failure while saving certificate %q [%v]", certFile, err)
		}
		if err = certutil.WriteKey(keyFile, certutil.EncodePrivateKeyPEM(key)); err != nil {
			return fmt.Errorf("failure while saving key %q [%v]",keyFile, err)
		}
		fmt.Printf("Generated cert %q.\n", certFile)
		fmt.Printf("Generated key %q.\n", keyFile)
	}
	return nil
}

// getAltNames builds an AltNames object for the certutil to use when generating the certificates
func getAltNames(cfgAltNames []string) certutil.AltNames {
	altNames := certutil.AltNames{}

	// Populate IPs/DNSNames from AltNames
	for _, altname := range cfgAltNames {
		if ip := net.ParseIP(altname); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		}
	}
	return altNames
}
