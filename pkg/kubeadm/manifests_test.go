package kubeadm

import (
	"testing"
)


func TestWriteManifests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	k, err:= getKubeadmCfg()
	if err != nil {
		t.Error(err)
	}

	// Simple case
	if err := k.WriteManifests(); err != nil {
		t.Error(err)
	}

	// Overwrite case
	if err := k.WriteManifests(); err != nil {
		t.Error(err)
	}
}
