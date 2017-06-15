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

// TODO: Add mockable interface for testing this package without reference to specific clientV3 lib

// Client represents an etcd client configuration.
type Client struct {
	Endpoints			string
	CaFileName			string
	ClientCertFileName	string
	ClientKeyFileName	string
}

// Clienter allows for mocking out this lib for testing
type Clienter interface {
	Get(key string) (value string, err error)
	GetOrCreateLock(key string) (mylock bool, err error)
	PutTx(key string, value string) (err error)
	Delete(key string) (err error)
}

// Verify the implementation here satisfies the abstract interface
var _ Clienter = (*Client)(nil)

var (
	// Timeout - For now a constant
	Timeout = 5 * time.Second

	// MaxTransactionTime - Also a constant - for now
	MaxTransactionTime = 120 * time.Second
)

// New creates a new etcd client from configuration
func New(cfg Client) *Client {
	return &cfg
}

// Get - Will return:
// - The the string value for a given key if present
// - Will return an err for all other occasions
func (c *Client) Get(key string) (value string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(*c, Timeout)
	if err != nil {
		log.Printf("Error getting client:%q", err)
		return "", err
	}
	defer cli.Close()

	log.Printf("Getting %q key value...", key)
	getresp, err := cli.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if len(getresp.Kvs) == 0 {
		return "", ErrKeyMissing
	}
	for _, ev := range getresp.Kvs {
		log.Debugf("%q values: key: %q = %q, version=%q\n", key, ev.Key, ev.Value, ev.Version)
		value = string(ev.Value[:])
		break
	}
	//log.Printf("%q key has specific value: %q\n", key, value)
	cancel() // context
	return value, err
}

// GetOrCreateLock obtains a lock (true) if the first client to create lock
// If TTL expired, will obtain lock (reset TTL)
// If TTL not expired will return false
func (c *Client) GetOrCreateLock(key string) (mylock bool, err error) {
	mylock = false

	err = c.SetLock(key)
	if err != nil {
		if err == ErrKeyAlreadyExists {
			log.Printf("Lock allready created...")
			// Need to check TTL and if required, transactionally re-create Lock..
			mylock, err = c.TryRecreateLock(key)
		}
	} else {
		log.Printf("Lock obtained...")
		mylock = true
	}
	return mylock, err
}

// SetLock create an ETCD lock key with a TTL from now
func (c *Client) SetLock(key string) (err error) {
	now := time.Now()
	ttl := now.Add(MaxTransactionTime)

	// Try and create lock item with value of TTL
	err = c.PutTx(key, ttl.Format(time.RFC3339))
	return err
}

// TryRecreateLock will recreate a Lock IF TTL of existing lock has expired.
// Returns true if lock obtained (re-created as TTL expired)
// Returns false if existing lock still valid
func (c *Client) TryRecreateLock(key string) (recreated bool, err error) {
	othersTTLString, err := c.Get(key)
	if err != nil {
		// Shouldn't get this unless terminal...
		log.Printf("Lock (key - %q) not obtained, Can't get key:%q", key, err)
		return false, err
	}
	// We have old TTL - parse it
	otherTTLTime, e := time.Parse(
		time.RFC3339,
		othersTTLString)
	if e != nil {
		// Error parsing lock, corrupt, overwrite and get lock
		err = c.OverWriteLock(key)
	} else {
		// See if TTL has passed and we should assume lock...
		if time.Now().After(otherTTLTime) {
			err = c.OverWriteLock(key)
		} else {
			log.Printf("Lock (key - %q) not obtained, TTL exists:%q", key, othersTTLString)
			return false, nil
		}
	}
	return false, err
}

// OverWriteLock will delete and re-create a lock
// TODO: this needs to be done as a transaction!
func(c *Client) OverWriteLock(key string) (err error) {
	err = c.Delete(key)
	if err != nil {
		log.Printf("Failed deleteing lock:%q", key)
	}
	err = c.SetLock(key)
	if err != nil {
		log.Printf("Failed creating lock:%q", key)
	}
	return err
}

// Delete - will remove a key from etcd
func (c *Client) Delete(key string) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(*c, Timeout)
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.Delete(ctx, key)
	cancel()
	return err
}

// PutTx - Puts with a transaction (will NOT create new revision)
// Will ensure only a single version is ever stored.
// Returns error if key already existed
func (c *Client) PutTx(key string, value string) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	cli, err := getEtcdClient(*c, Timeout)
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

func getEtcdClient(config Client, timeout time.Duration) (cli *clientv3.Client, err error) {

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
	}
	return cli, nil
}
