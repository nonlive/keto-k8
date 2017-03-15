package kmm

import (
	"os"
	"time"
	"errors"
	"log"

	"github.com/UKHomeOffice/kmm/pkg/etcd"
	"github.com/UKHomeOffice/kmm/pkg/kubeadm"
	"github.com/UKHomeOffice/kmm/pkg/fileutil"
)

const AssetKey string = "kmm-asset-key"
const AssetLockKey string = "kmm-asset-lock"

type Config struct {
	KubeadmCfg	kubeadm.Config
	KubeCaCert	string
	KubeCaKey	string
}

type KmmAssets struct {
	Value string
	Owner string
	CreatedAt time.Time
}

// kmm core logic
func GetAssets(cfg Config) (err error) {
	// First check for CA file...
	if _, err := os.Stat(cfg.KubeCaCert); os.IsNotExist(err) {
		return errors.New("Kube CA cert not found at:" + cfg.KubeCaCert)
	}
	if _, err := os.Stat(cfg.KubeCaKey); os.IsNotExist(err) {
		return errors.New("Kube CA key not found at:" + cfg.KubeCaKey)
	}
	if _, err := os.Stat(kubeadm.PkiDir); os.IsNotExist(err) {
		os.Mkdir(kubeadm.PkiDir, os.ModePerm)
	}

	err = fileutil.CopyFile(cfg.KubeCaCert, kubeadm.CaCertFile)
	if err != nil {
		return err
	}
	err = fileutil.SymlinkFile(cfg.KubeCaKey, kubeadm.CaKeyFile)
	if err != nil {
		return err
	}

	pkiCreated := false
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
				log.Printf("Obtained lock, creating assets...\n")

				// We can create the master assets here
 				err = kubeadm.CreatePKI(cfg.KubeadmCfg)
				if err == nil {
					pkiCreated = true
					log.Printf("Loading assets off disk...\n")
					assets, err = kubeadm.GetAssets(cfg.KubeadmCfg)
					if err == nil {
						err = etcd.PutTx(cfg.KubeadmCfg.EtcdClientConfig, AssetKey, assets)
					}
				}
				if err != nil {
					errC := CleanUp(cfg, true, false)
					if errC != nil {
						log.Printf("Error releasing Lock HELP!\n")
					}
					return err
				}
				// Do we unlock here? - no need, assets exist!
			} else {
				// We need to try and get the assets again after a back off
				time.Sleep(20 * time.Second)
			}
		} else if err != nil {
			return err
		} else {
			// Assest present in etcd so save assets
			log.Printf("Saving assets to disk...\n")
			if err := kubeadm.SaveAssets(cfg.KubeadmCfg, assets); err != nil {
				return err
			}
		}
	}
	// We have the shared assets, now re-create anything missing...
	if ! pkiCreated {
		if err := kubeadm.CreatePKI(cfg.KubeadmCfg) ; err != nil {
			return err
		}
	}
	err = kubeadm.CreateKubeConfig(cfg.KubeadmCfg)
	return err
}

func CleanUp(cfg Config, releaseLock bool, deleteAssets bool) (err error) {

	if releaseLock {
		log.Printf("Releasing lock...\n")
		if err = etcd.Delete(cfg.KubeadmCfg.EtcdClientConfig, AssetLockKey); err != nil {
			return err
		}
		log.Printf("Released lock\n")
	}
	if deleteAssets {
		log.Printf("Releasing assets...\n")
		if err = etcd.Delete(cfg.KubeadmCfg.EtcdClientConfig, AssetKey); err != nil {
			return err
		}
	}
	return nil
}
