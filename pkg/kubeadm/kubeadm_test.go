package kubeadm

//go:generate mockery -dir $GOPATH/src/github.com/UKHomeOffice/keto-k8/pkg/kubeadm -name=Kubeadmer

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/UKHomeOffice/keto-k8/pkg/etcd"
	"github.com/UKHomeOffice/keto-k8/pkg/kubeadm/pkiutil"
)

const pkiPath = "/etc/kubernetes/pki"
const pkiCaName = "kube-ca"

func TestCreatePKI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	k, err:= getKubeadmCfg()
	if err != nil {
		t.Error(err)
	}
	// Cleanup
	os.RemoveAll(pkiPath)
	os.MkdirAll(pkiPath, 0700)

	// Simple case - just run it
	if err:= k.CreatePKI(); err != nil {
		t.Error(err)
	}
}

func TestCreateKubeConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	k, err:= getKubeadmCfg()
	if err != nil {
		t.Error(err)
	}

	// Simple case
	if err:= k.CreateKubeConfig(); err != nil {
		t.Error(err)
	}
}

func TestLoadLoadAndSerializeAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	k, err:= getKubeadmCfg()
	if err != nil {
		t.Error(err)
	}

	// Simple case
	assets, err := k.LoadAndSerializeAssets()
	if err != nil {
		t.Error(err)
	}
	if assets == "" {
		t.Error(fmt.Errorf("expecting a serialized string of some sort, got an empty string"))
	}

}

func TestSaveAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	k, err:= getKubeadmCfg()
	if err != nil {
		t.Error(err)
	}

	// Simple case
	assets, err := k.LoadAndSerializeAssets()
	if err != nil {
		t.Error(err)
	}
	if err := k.SaveAssets(assets); err != nil {
		t.Error(err)
	}
}

func getKubeadmCfg() (*Config, error) {
	const k8versionfile = "../../k8version.cfg"

	if err := createCa(); err != nil {
		return nil, err
	}
	// Get the version from the version of kubeadm we'll put in docker file (and should have present)...
	// (normally comes from cloud provider)
	b, err := ioutil.ReadFile(k8versionfile)
	if err != nil {
		return nil, err
	}
	a := strings.Split(string(b), "=")
	if len(a) != 2 {
		return nil, fmt.Errorf("Invalid source string (missing '=' ) in file:%q", k8versionfile)
	}
	if a[0] != "K8S_VERSION" {
		return nil, fmt.Errorf("Invalid source string (missing K8_VERSION=) in file:%q", k8versionfile)
	}
	url, _ := url.Parse("https://localhost:6443")
	k8Version := strings.TrimSpace(a[1])
	return &Config{
		EtcdClientConfig: etcd.Client{
			Endpoints:  "https://127.0.0.1:2379",
			CaFileName: CaCertFile,

		},
		APIServer:		url,
		CaCert:			CaCertFile,
		CaKey:			CaKeyFile,
		CloudProvider:	"",
		KubeVersion:	k8Version,
		MasterCount:	1,
		PodNetworkCidr:	"",
	}, nil
}

func createCa() (err error) {
	// use cfssl for now as we use it internally for most CA's
	var (
		cert *x509.Certificate
		key  *rsa.PrivateKey
	)
	if cert, key, err = pkiutil.NewCertificateAuthority(); err != nil {
		return err
	}
	if err := pkiutil.WriteCertAndKey(pkiPath, pkiCaName, cert, key); err != nil {
		return err
	}
	return nil
}
