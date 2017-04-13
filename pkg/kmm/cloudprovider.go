package kmm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/UKHomeOffice/keto/pkg/cloudprovider"
)

type CloudAsset struct {
	FileName	string
	Value		[]byte
	Mode		os.FileMode
}

func SaveCloudAssets(cloudprovider, etcdCa, etcdCaKey, kubeCa, kubeCaKey string) (error) {
	node, err := getNodeInterface(cloudprovider)
	if err != nil {
		return err
	}
	assets, err := node.GetAssets()
	if err != nil {
		return err
	}
	var files = []CloudAsset{
		CloudAsset{
			FileName:	etcdCa,
			Value:		assets.EtcdCACert,
			Mode:		0644,
		},
		CloudAsset{
			FileName:	etcdCaKey,
			Value:		assets.EtcdCAKey,
			Mode:		0640,
		},
		CloudAsset{
			FileName:	kubeCa,
			Value:		assets.KubeCACert,
			Mode:		0644,
		},
		CloudAsset{
			FileName:	kubeCaKey,
			Value:		assets.KubeCAKey,
			Mode:		0640,
		},
	}
	for _, file := range files {
		if _, err := os.Stat(file.FileName); os.IsNotExist(err) {
			dir := filepath.Dir(file.FileName)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err = os.MkdirAll(dir, 0755); err != nil {
					return err
				}
			}

			// Only write a file if it didn't exist
			err = ioutil.WriteFile(file.FileName, file.Value, file.Mode)
			if err != nil {
				return fmt.Errorf("Cloud Asset [%q] could not saved [%v]", file.FileName, err)
			}
			log.Printf("Saved Cloud Asset [%q]", file.FileName)
		} else {
			log.Printf("Cloud Asset [%q] exists already", file.FileName)
		}
	}

	return nil
}

func getNodeInterface(cloudName string) (node cloudprovider.Node, err error) {
	var cloud cloudprovider.Interface
	if cloud, err = cloudprovider.InitCloudProvider(cloudName, nil); err != nil {
		return nil, err
	}
	var supported = false
	node, supported = cloud.Node()
	if supported {
		log.Printf("Cloud Provider Initialized [%q]", cloud.ProviderName())
	} else {
		return nil, fmt.Errorf("Cloud Provider set [%q] but node interface not supported", cloud.ProviderName())
	}
	return node, nil
}