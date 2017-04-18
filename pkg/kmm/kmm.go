package kmm

import (
	"os"
	"time"
	"errors"
	"net/url"
	"fmt"
	log "github.com/Sirupsen/logrus"

	"github.com/UKHomeOffice/keto-k8/pkg/etcd"
	"github.com/UKHomeOffice/keto-k8/pkg/kubeadm"
	"github.com/UKHomeOffice/keto-k8/pkg/fileutil"
	"github.com/UKHomeOffice/keto/pkg/cloudprovider"
)

const AssetKey string = "kmm-asset-key"
const AssetLockKey string = "kmm-asset-lock"

type Config struct {
	KubeadmCfg           kubeadm.Config

	/*
	 To provide the least change to kubeadm and prevent access to the key,
	 we copy and link from the persistent source as appropriate

	 TODO: Update kubeadm to allow for specified CA key locations
	  */
	KubePersistentCaCert string
	KubePersistentCaKey  string
}

type KmmAssets struct {
	Value string
	Owner string
	CreatedAt time.Time
}

func Manifests(cfg kubeadm.Config) (err error) {
	if err = kubeadm.WriteManifests(cfg); err != nil {
		return err
	}
	return nil
}

// kmm core logic
func GetAssets(cfg Config) (err error) {

	if err = updateCloudCfg(&cfg) ; err != nil {
		return err
	}
	if err = copyKubeCa(cfg) ; err != nil {
		return err
	}
	if err = kubeadm.WriteManifests(cfg.KubeadmCfg); err != nil {
		return err
	}

	bootStrappedHere := false
	assets := ""

	// Keep trying to get Assets
	for assets == "" {
		assets, err = etcd.Get(cfg.KubeadmCfg.EtcdClientConfig, AssetKey)
		if err == etcd.ErrKeyMissing {
			log.Printf("Assets not present in etcd...\n")
			// obtain lock...
			mylock, err := etcd.GetLock(cfg.KubeadmCfg.EtcdClientConfig, AssetLockKey)
			if err != nil {
				// May need to add retry logic?
				return err
			}
			if mylock {
				log.Printf("Obtained lock, creating assets...")
				if assets, err = bootstrapOnce(cfg); err != nil {
					return err
				} else {
					bootStrappedHere = true
					// Only share assets when all done OK!
					if err = etcd.PutTx(cfg.KubeadmCfg.EtcdClientConfig, AssetKey, assets) ; err != nil {
						return err
					}
				}
			} else {
				// We need to try and get the assets again after a back off
				time.Sleep(20 * time.Second)
			}
		} else if err != nil {
			return err
		} else {
			// Assets present in etcd so save assets
			log.Printf("Saving assets to disk...")
			if err := kubeadm.SaveAssets(cfg.KubeadmCfg, assets); err != nil {
				return err
			}
		}
	}
	// We have the shared assets, now re-create anything missing...
	if ! bootStrappedHere {
		if err := kubeadm.CreatePKI(cfg.KubeadmCfg) ; err != nil {
			return err
		}
		if err = kubeadm.CreateKubeConfig(cfg.KubeadmCfg) ; err != nil {
			return err
		}
	}
	return nil
}

func CleanUp(cfg Config, releaseLock bool, deleteAssets bool) (err error) {

	if releaseLock {
		log.Printf("Releasing lock...")
		if err = etcd.Delete(cfg.KubeadmCfg.EtcdClientConfig, AssetLockKey); err != nil {
			return err
		}
		log.Printf("Released lock")
	}
	if deleteAssets {
		log.Printf("Releasing assets...")
		if err = etcd.Delete(cfg.KubeadmCfg.EtcdClientConfig, AssetKey); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapOnce(cfg Config) (assets string, err error) {

	defer CleanUp(cfg, true, false)

	// We can create the master assets here
	if err = kubeadm.CreatePKI(cfg.KubeadmCfg); err != nil {
		return "", err
	}
	log.Printf("Loading assets off disk...")
	assets, err = kubeadm.GetAssets(cfg.KubeadmCfg)

	// We have the assets but we must NOT proceed until we've finish bootstrapping / sharing...
	if err = kubeadm.CreateKubeConfig(cfg.KubeadmCfg); err != nil {
		return "", err
	}
	if err = kubeadm.Addons(cfg.KubeadmCfg); err != nil {
		return "", err
	}
	return assets, nil
}

// Copy Kube CA and link CA key to kubeadm expected locations (if not there already)
func copyKubeCa(cfg Config) (err error) {
	// First check for CA file...
	if _, err := os.Stat(cfg.KubePersistentCaCert); os.IsNotExist(err) {
		return errors.New("Kube CA cert not found at:" + cfg.KubePersistentCaCert)
	}
	if _, err := os.Stat(cfg.KubePersistentCaKey); os.IsNotExist(err) {
		return errors.New("Kube CA key not found at:" + cfg.KubePersistentCaKey)
	}
	if _, err = os.Stat(kubeadm.PkiDir); os.IsNotExist(err) {
		os.Mkdir(kubeadm.PkiDir, os.ModePerm)
	}

	err = fileutil.CopyFile(cfg.KubePersistentCaCert, kubeadm.CaCertFile)
	if err != nil {
		return err
	}
	err = fileutil.SymlinkFile(cfg.KubePersistentCaKey, kubeadm.CaKeyFile)
	if err != nil {
		return err
	}
	return nil
}

// Update config based on cloud provider, if specified
func updateCloudCfg(cfg *Config) (err error) {
	// Now get the cloud provider to get the kubeapi url and k8 version:
	if cfg.KubeadmCfg.CloudProvider != "" {
		var node cloudprovider.Node
		if node, err = getNodeInterface(cfg.KubeadmCfg.CloudProvider); err != nil {
			return err
		}
		var api string
		if api, err = node.GetKubeAPIURL(); err != nil {
			return fmt.Errorf("Error getting Api server from cloud provider:%q", err)
		}
		// TODO: detect if a port set here...
		url, err := url.Parse(api + ":6443")
		if err != nil {
			return fmt.Errorf("Error parsing Api server %s [%v]", api, err)
		} else {
			if len(api) > 0 {
				cfg.KubeadmCfg.ApiServer = url
			} else {
				// url.Parse seems to always parse without error!
				return fmt.Errorf("Empty API server [%s] obtained from cloud provider", api)
			}
		}
		if cfg.KubeadmCfg.KubeVersion, err = node.GetKubeVersion(); err != nil {
			return fmt.Errorf("Kubernetes version not specified from cloud provider [%v]", err)
		} else {
			if len(cfg.KubeadmCfg.KubeVersion) == 0 {
				return fmt.Errorf("Error parsing Api server %s", api)
			}
		}
	}

	return nil
}
