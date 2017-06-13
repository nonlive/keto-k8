package kmm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/UKHomeOffice/keto-k8/pkg/constants"
	"github.com/UKHomeOffice/keto-k8/pkg/fileutil"
	"github.com/coreos/go-systemd/dbus"
)

// CreateAndStartKubelet will create Kubelet
func CreateAndStartKubelet(cloudProvider, kubeVersion string, master bool) error {

	// Render kubelet.service
	data := struct {
		CloudProviderName string
		KubeVersion       string
		IsMaster          bool
	}{
		CloudProviderName: cloudProvider,
		KubeVersion:       kubeVersion,
		IsMaster:          master,
	}
	t := template.Must(template.New("kubeletUnit").Parse(kubeletTemplate))
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return fmt.Errorf("Error generating kubelet unit [%v] from template:\n%v", err, kubeletTemplate)
	}

	// Get D-bus connection
	target := path.Base(constants.KubeletUnitFileName)
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Manage unit file
	if fileutil.ExistFile(constants.KubeletUnitFileName) {
		// Tidy up existing file...
		oldUnit, err := ioutil.ReadFile(constants.KubeletUnitFileName)
		if err != nil {
			return fmt.Errorf("Error [%v] reading existing unit [%v]", err, kubeletTemplate)
		}
		if string(oldUnit) != b.String() {
			// delete file
			if err := os.Remove(constants.KubeletUnitFileName); err != nil {
				return fmt.Errorf("Error [%v] removing existing kubelet unit [%v]",
					err,
					constants.KubeletUnitFileName)
			}
			// TODO: stop unit if already running
		} else {
			// TODO: return IF kubelet already running, return here (no change)

		}
	}
	if !fileutil.ExistFile(constants.KubeletUnitFileName) {
		// Create unit
		if err := ioutil.WriteFile(constants.KubeletUnitFileName, []byte(b.Bytes()), 0644); err != nil {
			return fmt.Errorf("Can not save unit file [%v]: [%v]",
				constants.KubeletUnitFileName,
				err)
		}
	}
	// Daemon-reload TODO: make reload unit specific
	if err := conn.Reload(); err != nil {
		return fmt.Errorf("problem reloading systemd units after adding %q; [%v]", target, err)
	}

	// Start / restart unit
	reschan := make(chan string)
	if _, err := conn.StartUnit(target, "replace", reschan); err != nil {
		return fmt.Errorf("can't start unit [%v] - [%v]", target, err)
	}
	job := <-reschan
	if job != "done" {
		return fmt.Errorf("unknown error starting [%v]", target)
	}

	// TODO: enable unit (link if required)
	return nil
}
