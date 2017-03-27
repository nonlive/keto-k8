package kubeadm

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os/exec"
	"io/ioutil"
	"strings"

	certutil "github.com/UKHomeOffice/kmm/pkg/client-go/util/cert"
	kubeadmconstants "github.com/UKHomeOffice/kmm/pkg/kubeadm/constants"
	log "github.com/Sirupsen/logrus"

	"github.com/UKHomeOffice/kmm/pkg/kubeadm/pkiutil"
	"github.com/UKHomeOffice/kmm/pkg/etcd"
)

const CmdKubeadm string = "kubeadm"

var (
	CmdOptsCerts 		= []string {"alpha", "phase", "certs", "selfsign", "--cert-altnames"}
	CmdOptsKubeconfig 	= []string {"alpha", "phase", "kubeconfig", "client-certs"}
	PkiDir string 		= kubeadmconstants.KubernetesDir + "/pki"
	CaCertFile string	= kubeadmconstants.KubernetesDir + "/pki" + "/" + kubeadmconstants.CACertAndKeyBaseName + ".crt"
	CaKeyFile string 	= kubeadmconstants.KubernetesDir + "/pki" + "/" + kubeadmconstants.CACertAndKeyBaseName + ".key"
)

// represents runtime params cfg structure.
type Config struct {
	EtcdClientConfig	etcd.ClientConfig
	CaCert				string
	CaKey				string
	ApiServer			string
	KubeletId			string
}

type SharedAssets struct {
	FrontProxyCa	string
	FrontProxyCaKey	string
	SaPub			string
	SaKey			string
}

// Must grab any assets off disk
// Return an error if there are no assets (and empty string)
func GetAssets(cfg Config) (assets string, err error) {
	assets = ""

	var saPub *rsa.PublicKey
	var saKey *rsa.PrivateKey
	saKey, err = pkiutil.TryLoadKeyFromDisk(PkiDir, kubeadmconstants.ServiceAccountKeyBaseName)
	if err != nil {
		return "", fmt.Errorf("SA private key could not be loaded properly [%v]", err)
	}
	saPub, err = pkiutil.TryLoadPublicKeyFromDisk(PkiDir, kubeadmconstants.ServiceAccountKeyBaseName)
	if err != nil {
		return "", fmt.Errorf("SA public key could not be loaded properly [%v]", err)
	}

	var frontProxyCACert *x509.Certificate
	var frontProxyCAKey *rsa.PrivateKey
	frontProxyCACert, frontProxyCAKey, err = pkiutil.TryLoadCertAndKeyFromDisk(PkiDir, kubeadmconstants.FrontProxyCACertAndKeyBaseName)
	if err != nil || frontProxyCACert == nil || frontProxyCAKey == nil {
		return "", fmt.Errorf("Front proxy certificate and/or key existed but they could not be loaded properly")
	}

	// The certificate and key could be loaded, but the certificate is not a CA
	if !frontProxyCACert.IsCA {
		return "", fmt.Errorf("certificate and key could be loaded but the certificate is not a CA")
	}

	saPubPemBytes, _ := certutil.EncodePublicKeyPEM(saPub)
	// Re-encode the values now we've checked them...
	sharedAssets := &SharedAssets{
		SaPub:				string(saPubPemBytes[:]),
		SaKey:				string(certutil.EncodePrivateKeyPEM(saKey)[:]),
		FrontProxyCa:		string(certutil.EncodeCertPEM(frontProxyCACert)[:]),
		FrontProxyCaKey:	string(certutil.EncodePrivateKeyPEM(frontProxyCAKey)[:]),
	}

	// Now json encode the structure
	assetsBytes, _ := json.Marshal(sharedAssets)
	assets = string(assetsBytes)

	return assets, nil
}

func SaveAssets(cfg Config, assets string) (err error) {
	pkiDir := PkiDir + "/"
	sharedAssets := SharedAssets{}
	json.Unmarshal([]byte(assets), &sharedAssets)

	// Now save each of the pem files...
	err = ioutil.WriteFile(pkiDir + kubeadmconstants.ServiceAccountPublicKeyName, []byte(sharedAssets.SaPub), 0644)
	if err != nil {
		return fmt.Errorf("Service Account public key could not saved [%v]", err)
	}
	err = ioutil.WriteFile(pkiDir + kubeadmconstants.ServiceAccountPrivateKeyName, []byte(sharedAssets.SaKey), 0600)
	if err != nil {
		return fmt.Errorf("Service Account private key could not saved [%v]", err)
	}
	err = ioutil.WriteFile(pkiDir + kubeadmconstants.FrontProxyCACertName, []byte(sharedAssets.FrontProxyCa), 0644)
	if err != nil {
		return fmt.Errorf("Front proxy public ca cert could not saved [%v]", err)
	}
	err = ioutil.WriteFile(pkiDir + kubeadmconstants.FrontProxyCAKeyName, []byte(sharedAssets.FrontProxyCaKey), 0600)
	if err != nil {
		return fmt.Errorf("Front proxy private key could not saved [%v]", err)
	}

	return nil
}

// Create all PKI assests on disk
func CreatePKI(cfg Config) (err error) {
	args := append(CmdOptsCerts, cfg.ApiServer)
	kubeadmOut, err := runKubeadm(cfg, args)
	log.Printf("Output:\n" + kubeadmOut)
	return err
}

func CreateKubeConfig(cfg Config) (err error) {
	if err = createAKubeCfg(cfg, kubeadmconstants.AdminKubeConfigFileName,
		"kubernetes-admin", kubeadmconstants.MastersGroup); err !=nil {

		return err
	}
	if err = createAKubeCfg(cfg, kubeadmconstants.KubeletKubeConfigFileName,
		"system:node:" + cfg.KubeletId, kubeadmconstants.NodesGroup); err !=nil {

		return err
	}
	if err = createAKubeCfg(cfg, kubeadmconstants.ControllerManagerKubeConfigFileName,
		kubeadmconstants.ControllerManagerUser, ""); err !=nil {

		return err
	}
	if err = createAKubeCfg(cfg, kubeadmconstants.SchedulerKubeConfigFileName,
		kubeadmconstants.SchedulerUser, ""); err !=nil {
		return err
	}
	return nil
}

// Run kubeadm to create a kubeconfig file...
func createAKubeCfg(cfg Config, file string, cn string, org string) (err error) {
	args := append(CmdOptsKubeconfig,
		"--client-name ", cn,
		"--server", cfg.ApiServer)

	if len(org) > 0 {
		args = append(args,
			"--organization ", org)
	}

	kubecfgContents, err :=	runKubeadm(cfg, args)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(PkiDir + "/" + file, []byte(kubecfgContents), 0600)
	return err
}

func runKubeadm(cfg Config, cmdArgs []string) (out string, err error) {
	var cmdOut []byte

	cmdName := CmdKubeadm
	log.Printf("Running:%v %v", cmdName, strings.Join(cmdArgs, " "))
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).CombinedOutput(); err != nil {
		return string(cmdOut[:]), err
	}
	return string(cmdOut[:]), nil
}
