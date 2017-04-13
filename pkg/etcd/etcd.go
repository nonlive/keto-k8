package etcd

import (
	"time"
	"strings"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/pkg/transport"
)

// TODO: rewrite a client interface (for testability)

// Config represents a controller configuration.
type ClientConfig struct {
	Endpoints			string
	CaFileName			string
	ClientCertFileName	string
	ClientKeyFileName	string
}

var (
	Timeout = 5 * time.Second
	MaxTransactionTime = 120 * time.Second
)

// Will return:
// - The the string value for a given key if present
// - Will return an err for all other occasions
func Get(cfg ClientConfig, key string) (value string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(cfg, Timeout)
	if err != nil {
		log.Printf("Error getting client:%q", err)
		return "", err
	}
	defer cli.Close()

	log.Printf("Getting %q key value...", key)
	getresp, err := cli.Get(ctx, key)
	if err != nil {
		return "", err
	} else {
		if len(getresp.Kvs) == 0 {
			return "", ErrKeyMissing
		}
		for _, ev := range getresp.Kvs {
			log.Debugf("%q values: key: %q = %q, version=%q\n", key, ev.Key, ev.Value, ev.Version)
			value = string(ev.Value[:])
			break
		}
	}
	//log.Printf("%q key has specific value: %q\n", key, value)
	cancel() // context
	return value, err
}

// Will obtain a lock (true) if the first client to create lock
// If TTL expired, will obtain lock (reset TTL)
// If TTL not expired will return false
func GetLock(cfg ClientConfig, key string) (mylock bool, err error) {
	mylock = false

	err = setLock(cfg, key)
	if err != nil {
		if err == ErrKeyAlreadyExists {
			log.Printf("Lock allready created...")
			// Need to check TTL and if required, transactionally re-create Lock..
			mylock, err = tryRecreateLock(cfg, key)
		}
	} else {
		log.Printf("Lock obtained...")
		mylock = true
	}
	return mylock, err
}

func setLock(cfg ClientConfig, key string) (err error) {
	now := time.Now()
	ttl := now.Add(MaxTransactionTime)

	// Try and create lock item with value of TTL
	err = PutTx(cfg, key, ttl.Format(time.RFC3339))
	return err
}

// Will recreate a Lock IF TTL of existing lock has expired.
// Returns true if lock obtained (re-created as TTL expired)
// Returns falue if existing lock still valid
func tryRecreateLock(cfg ClientConfig, key string) (recreated bool, err error) {
	othersTtlString, err := Get(cfg, key)
	if err != nil {
		// Shouldn't get this unless terminal...
		log.Printf("Lock (key - %q) not obtained, Can't get key:%q", key, err)
		return false, err
	} else {
		// We have old TTL - parse it
		otherTtlTime, e := time.Parse(
			time.RFC3339,
			othersTtlString)
		if e != nil {
			// Error parsing lock, corrupt, overwrite and get lock
			err = overWriteLock(cfg, key)
		} else {
			// See if TTL has passed and we should assume lock...
			if time.Now().After(otherTtlTime) {
				err = overWriteLock(cfg, key)
			} else {
				log.Printf("Lock (key - %q) not obtained, TTL exists:%q", key, othersTtlString)
				return false, nil
			}
		}
	}
	return false, err
}

func overWriteLock(cfg ClientConfig, key string) (err error) {
	err = Delete(cfg, key)
	if err != nil {
		log.Printf("Failed deleteing lock:%q", key)
	}
	err = setLock(cfg, key)
	if err != nil {
		log.Printf("Failed creating lock:%q", key)
	}
	return err
}

func Delete(cfg ClientConfig, key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(cfg, Timeout)
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.Delete(ctx, key)
	cancel()
	return err
}

// Puts with a transaction (will NOT create new revision)
// Will ensure only a single version is ever stored.
// Returns error if key already existed
func PutTx(cfg ClientConfig, key string, value string) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(cfg, Timeout)
	if err != nil {
		return err
	}
	defer cli.Close()

	kvc := clientv3.NewKV(cli)

	// perform a put only if key is missing
	// It is useful to do the check (transactionally) to avoid overwriting
	// the existing key which would generate potentially unwanted events,
	// unless of course you wanted to do an overwrite no matter what.
	txRet, err := kvc.Txn(ctx).
	If(clientv3util.KeyMissing(key)).
	Then(clientv3.OpPut(key, value)).
	Commit()

	cancel() // context

	if ! txRet.Succeeded {
		// We didn't create the lock - indicate with dedicated error:
		log.Printf("Transaction didn't succeed - we didn't create lock!")
		err = ErrKeyAlreadyExists
	} else {
		log.Debugf("Created item:%q...", value)
	}
	return err
}

func getEtcdClient(config ClientConfig, timeout time.Duration) (cli *clientv3.Client, err error) {

	endPoints := strings.Split(config.Endpoints, ",")
	cfg := clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: timeout,
	}
	if config.CaFileName == "" {
		log.Printf("No ca file specified. not using client certs")
	} else {
		tlsInfo := transport.TLSInfo{
			CertFile: config.ClientCertFileName,
			KeyFile:  config.ClientKeyFileName,
			CAFile:   config.CaFileName,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = tlsConfig
	}
	cli, err = clientv3.New(cfg)
	if err != nil {
		return nil, err
	} else {
		return cli, nil
	}
}
