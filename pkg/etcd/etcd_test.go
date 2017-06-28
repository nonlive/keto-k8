package etcd

//go:generate mockery -dir $GOPATH/src/github.com/UKHomeOffice/keto-k8/pkg/etcd -name=Clienter

import (
	"fmt"
	"log"
	"path"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/UKHomeOffice/keto-k8/pkg/fileutil"
)

const containerName string = "ectd_int_test"

func TestMain(m *testing.M) {
	startEtcd3()
	ret := m.Run()
	stopEtcd3()
	os.Exit(ret)
}

func TestNew(t *testing.T) {
	if client := New(getClientCfg()); client == nil {
		t.Error(fmt.Errorf("Expected a valid client but got nil"))
	}
}

func TestGet(t *testing.T) {
	const testGetKey string = "testget"
	const testGetValue string = "value"

	if testing.Short() {
		t.Skip("skipping integration test")
	}
	e := getETCDClient()

	// Cleanup
	_ = e.Delete(testGetKey)

	if _, err := e.Get("nonexistingkey"); err != ErrKeyMissing {
		t.Error(fmt.Errorf("expected error %q but got %q", ErrKeyMissing, err))
	}
	if err := e.PutTx(testGetKey, testGetValue); err != nil {
		t.Error(fmt.Errorf("expected no error but got %q", err))
	}
	if value, err := e.Get(testGetKey); err != nil {
		t.Error(fmt.Errorf("expected no error but got %q", err))
	} else {
		if value != testGetValue {
			t.Error("expected %q when getting %q but got %q", testGetValue, testGetKey, value)
		}
	}
}

func TestDelete(t *testing.T) {
	const testDeleteKey string = "testdelete"
	const testDeleteValue string = "valuegobyebye"

	if testing.Short() {
		t.Skip("skipping integration test")
	}
	e := getETCDClient()

	// Cleanup
	_ = e.Delete(testDeleteKey)

	// normal test case (delete existing key)
	if err := e.PutTx(testDeleteKey, testDeleteValue); err != nil {
		t.Error(fmt.Errorf("expected no error but got %q", err))
	}
	if err := e.Delete(testDeleteKey); err != nil {
		t.Error(fmt.Errorf("expected no error but got %q", err))
	}
	if _, err := e.Get(testDeleteKey); err != ErrKeyMissing {
		t.Error(fmt.Errorf("expected key missing error but got %q", err))
	}

	// key doesn't exist case
	// TODO: need to parse the delete response in the lib
	// for now the delete with no error is expected by the existing use cases (luckly)
	// we should still parse the delete response and only accept this use case and no other
	if err := e.Delete(testDeleteKey); err != nil {
		t.Error(fmt.Errorf("expected no error but got error: %q", err))
	}
}

func TestPutTx(t *testing.T) {
	const testPutTxKey string = "testputtx"
	const testPutTxValue string  = "valueofaputtx"

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Simple case, non-existing key
	e := getETCDClient()

	// Cleanup
	_ = e.Delete(testPutTxKey)

	if err := e.PutTx(testPutTxKey, testPutTxValue); err != nil {
		t.Error(fmt.Errorf("did not manage PutTx() got error:%q", err))
	} else {
		if value, _ := e.Get(testPutTxKey); value != testPutTxValue {
			t.Error(fmt.Errorf("expected %q but got %q", testPutTxValue, ))
		}
	}

	// Existing key
	if err := e.PutTx(testPutTxKey, "testOtherValue"); err != ErrKeyAlreadyExists {
		t.Error(fmt.Errorf("did not get expected error, got:%q", err))
	} else {
		if value, _ := e.Get(testPutTxKey); value != testPutTxValue {
			t.Error(fmt.Errorf("expected %q but got %q", testPutTxValue, ))
		}
	}

	// TODO: some how test the transaction fail (not key exists before but during!)
}

func TestGetOrCreateLock(t *testing.T) {
	const testGetOrCreateLockKey string = "testgetorcreatelock"
	var testLongGetOrCreateLockTTL time.Duration = 120 * time.Second
	var testShortGetOrCreateLockTTL time.Duration = 1 * time.Second

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Simple case, non-existing lock
	e := getETCDClient()
	if lock, err := e.GetOrCreateLock(testGetOrCreateLockKey, testLongGetOrCreateLockTTL); err != nil {
		t.Error(fmt.Errorf("did not get lock result when expected got error:%q", err))
	} else {
		if !lock {
			t.Error(fmt.Errorf("expected lock == true"))
		}
	}

	// Test when lock has already been created (within TTL)
	if lock, err := e.GetOrCreateLock(testGetOrCreateLockKey, testLongGetOrCreateLockTTL); err != nil {
		t.Error(fmt.Errorf("did not get lock result when expected got error:%q", err))
	} else {
		if lock {
			t.Error(fmt.Errorf("expected lock == false"))
		}
	}

	// test when lock has been created but has expired (outside TTL)
	_ = e.Delete(testGetOrCreateLockKey)
	_, _ = e.GetOrCreateLock(testGetOrCreateLockKey, testShortGetOrCreateLockTTL)
	time.Sleep(testShortGetOrCreateLockTTL)
	if lock, err := e.GetOrCreateLock(testGetOrCreateLockKey, testShortGetOrCreateLockTTL); err != nil {
		t.Error(fmt.Errorf("did not get lock result when expected got error:%q", err))
	} else {
		if !lock {
			t.Error(fmt.Errorf("expected lock == true"))
		}
	}
	_ = e.Delete(testGetOrCreateLockKey)

	// test when lock is corrupted i.e. invalid (created with wrong version???)
	e.PutTx(testGetOrCreateLockKey, "not a good ttl!")
	if lock, err := e.GetOrCreateLock(testGetOrCreateLockKey, testShortGetOrCreateLockTTL); err != nil {
		t.Error(fmt.Errorf("did not get lock result when expected got error:%q", err))
	} else {
		if !lock {
			t.Error(fmt.Errorf("expected lock == true"))
		}
	}
}

func getETCDClient() *Client {
	return New(getClientCfg())
}

func startEtcd3() {
	script := getPath("tests/start_etcd_server.sh")
	if !fileutil.ExistFile(script) {
		log.Fatal(fmt.Errorf("Missing script - %q", script))
	}
	log.Printf("Running:%v", script)
	// This should always cleanup any data at the start of testing
	cmd := exec.Command(script, containerName, "cleanup")
	stdoutStderr, err := cmd.CombinedOutput()
	log.Printf("etcd output:\n%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
}

func stopEtcd3() {
	// transiently try and stop container...
	_, _ = exec.Command("docker", "stop", containerName).CombinedOutput()
}

func getClientCfg() Client {
	return Client{
		CaFileName:		getPath("tests/certs/ca.pem"),
		ClientCertFileName:	getPath("tests/certs/client.pem"),
		ClientKeyFileName:	getPath("tests/certs/client-key.pem"),
		Endpoints:		"https://127.0.0.1:2379/",
	}
}

func getPath(relPath string) (string) {
	wdPath, _ := os.Getwd()
	return path.Clean(wdPath + "/../../" + relPath)
}
