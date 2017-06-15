package kmm

//go:generate mockery -dir $GOPATH/src/github.com/UKHomeOffice/keto-k8/pkg/kmm -name=Interface

import (
	"fmt"
	"testing"
	"time"

	"github.com/UKHomeOffice/keto-k8/pkg/etcd"
	etcdMocks "github.com/UKHomeOffice/keto-k8/pkg/etcd/mocks"
	kmmMocks "github.com/UKHomeOffice/keto-k8/pkg/kmm/mocks"
	kubeadmMocks "github.com/UKHomeOffice/keto-k8/pkg/kubeadm/mocks"
)

const testAssets = "{}"

// testMock used for mockable interface
type testMock struct {
	Etcd    *etcdMocks.Clienter
	Kubeadm *kubeadmMocks.Kubeadmer
	Kmm     *kmmMocks.Interface
}

func getTestMock() (*testMock, *Config) {
	m := &testMock{
		Etcd:    &etcdMocks.Clienter{},
		Kubeadm: &kubeadmMocks.Kubeadmer{},
		Kmm:     &kmmMocks.Interface{},
	}

	kmm := &Config{}
	// Must exit tests!
	kmm.ExitOnCompletion = true
	kmm.Etcd = m.Etcd
	kmm.Kubeadm = m.Kubeadm
	kmm.Kmm = m.Kmm
	kmm.MasterBackOffTime = (time.Microsecond * 100)
	return m, kmm
}

func AddBootstapOnceAssertions(m *testMock) {
	m.Kubeadm.On("CreatePKI").Return(nil).Once()
	m.Kubeadm.On("LoadAndSerializeAssets").Return(testAssets, nil)
	m.Kubeadm.On("CreateKubeConfig").Return(nil).Once()
	m.Kmm.On("CreateAndStartKubelet", true).Return(nil).Once()

	// Note: Addons will call the same underlying kubeadmapi UpdateMasterRoleLabelsAndTaints
	m.Kubeadm.On("Addons").Return(nil).Once()
	m.Kmm.On("InstallNetwork").Return(nil).Once()
	m.Kmm.On("TokensDeploy").Return(nil).Once()
}

func AddMasterAssertions(m *testMock, primary bool) {
	// Methods we expect to always be called on masters:
	m.Kmm.On("UpdateCloudCfg").Return(nil)
	m.Kmm.On("CopyKubeCa").Return(nil)
	m.Kubeadm.On("WriteManifests").Return(nil)

	if primary {
		AddBootstapOnceAssertions(m)
	} else {
		m.Kubeadm.On("CreatePKI").Return(nil).Once()
		m.Kubeadm.On("CreateKubeConfig").Return(nil).Once()
		m.Kmm.On("CreateAndStartKubelet", true).Return(nil).Once()
		m.Kubeadm.On("UpdateMasterRoleLabelsAndTaints").Return(nil).Once()
	}
}

func TestBootStrappedOnce(t *testing.T) {
	m, k := getTestMock()

	AddBootstapOnceAssertions(m)

	if assets, err := k.BootstrapOnce(); err != nil {
		t.Error(err)
	} else {
		if assets != testAssets {
			t.Error(fmt.Errorf("Test assets not equal"))
		}
	}
	m.Kmm.AssertExpectations(t)
	m.Kubeadm.AssertExpectations(t)
}

func TestCreateOrGetSharedAssets(t *testing.T) {

	m, k := getTestMock()

	// Test primary master:
	// No assets stored, No pre-existing etcd lock, clean run...
	m.Etcd.On("Get", assetKey).Return("", etcd.ErrKeyMissing).Once()
	m.Etcd.On("GetOrCreateLock", assetLockKey).Return(true, nil).Once()
	m.Etcd.On("PutTx", assetKey, testAssets).Return(nil)

	AddMasterAssertions(m, true)

	if err := k.CreateOrGetSharedAssets(); err != nil {
		t.Error(err)
	}

	m.Kmm.AssertExpectations(t)
	m.Etcd.AssertExpectations(t)
	m.Kubeadm.AssertExpectations(t)
}

func TestCreateOrGetSharedAssetsSecondaryMaster(t *testing.T) {

	m, k := getTestMock()

	// Test secondary master:
	// Assets pre-stored, clean run...
	m.Etcd.On("Get", assetKey).Return(testAssets, nil).Once()
	m.Kubeadm.On("SaveAssets", testAssets).Return(nil).Once()

	// Assing expected outcomes from the secondary master
	AddMasterAssertions(m, false)

	if err := k.CreateOrGetSharedAssets(); err != nil {
		t.Error(err)
	}

	m.Kmm.AssertExpectations(t)
	m.Etcd.AssertExpectations(t)
	m.Kubeadm.AssertExpectations(t)
}
