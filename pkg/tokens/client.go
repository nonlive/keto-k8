package tokens

import (
	"io/ioutil"

	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"github.com/UKHomeOffice/keto-k8/pkg/constants"
)

// WriteKetoTokenEnv will write details needed by keto-tokens
func WriteKetoTokenEnv(cloud, apiURL string) (error) {

	envFileContents := "KETO_TOKENS_IMAGE=" + constants.KetoTokenImage + "\n" +
	                   "KETO_TOKENS_CLOUD=" + cloud + "\n" +
					   "KETO_TOKENS_TAG=" + constants.KetoTokenTagName + "\n" +
	                   "KETO_TOKENS_KUBELET_CONF=" + kubeadmconstants.KubernetesDir + "/bootstrap-kubelet.conf" + "\n" +
	                   "KETO_TOKENS_API_URL=" + apiURL + "\n"

	if err := ioutil.WriteFile(constants.KetoTokenEnvName, []byte(envFileContents), 0644); err != nil {
		return err
	}
	return nil
}
