package copy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var environmentCMD = &cobra.Command{
	Use:   "environment",
	Short: "Copy env variables configmap for a microservice in one environment to another environment",
	Long: `
	Copy env variables configmap for a microservice in one environment to another environment.

	go run main.go tools microservice copy environment \
		--application cde2e951-d40a-3548-8b45-64c0ded97940 \
		--microservice-name frontend \
		--from-env test \
		--to-env prod
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()

		sourceMicroserviceName := viper.GetString("tools.microservice.copy.environment.microservice-name")
		sourceEnvironment := viper.GetString("tools.microservice.copy.environment.from-env")
		destinationEnvironment := viper.GetString("tools.microservice.copy.environment.to-env")
		application := viper.GetString("tools.microservice.copy.environment.application")
		sourceConfigMapName := fmt.Sprintf("%s-%s-env-variables", sourceEnvironment, sourceMicroserviceName)
		namespace := fmt.Sprintf("application-%s", application)
		logContextWithMeta := logContext.WithFields(logrus.Fields{
			"k8sNamespace": namespace,
			"k8sConfigmap": sourceConfigMapName,
		})

		logContextWithMeta.WithField("microserviceName", sourceMicroserviceName).
			Info("Begins to copy env var configmap")

		kubeCtlGetCfgMap := exec.Command("kubectl", "-n", namespace, "get", "configmap", sourceConfigMapName, "-o", "json")
		_, err := kubeCtlGetCfgMap.StderrPipe()
		if err != nil {
			logContext.Fatal(err)
			return
		}

		configMapJson, err := kubeCtlGetCfgMap.Output()
		if err != nil {
			logContextWithMeta.WithField("error", err).
				Fatal("Failed to get configmap")
			return
		}

		var modifiedConfigMap map[string]interface{}
		json.Unmarshal(configMapJson, &modifiedConfigMap)
		destinationConfigMap := fmt.Sprintf("%s-%s-env-variables", destinationEnvironment, sourceMicroserviceName)
		metadata := modifiedConfigMap["metadata"].(map[string]interface{})
		metadata["name"] = destinationConfigMap
		// delete uid and version to ensure it patches
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")
		out, _ := json.Marshal(modifiedConfigMap)

		kubectlApply := exec.Command("kubectl", "-o", "json", "apply", "-f", "-")
		kubectlApply.Stdin = bytes.NewReader(out)
		kubectlApply.Stderr = os.Stderr

		kubectlApply.Start()
		kubeCtlGetCfgMap.Run()
		kubectlApply.Wait()

		logContextWithMeta.WithField("k8sDestinationConfigMap", destinationConfigMap).
			Info("Successfully copied configmap for env variables")
	},
}
