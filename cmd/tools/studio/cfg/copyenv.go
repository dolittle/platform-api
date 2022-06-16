package cfg

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

var heyCMD = &cobra.Command{
	Use:   "copyenv",
	Short: "Copy env variables configmap for a microservice in one environment to another environment",
	Long: `
	Copy env variables configmap for a microservice in one environment to another environment.

	go run main.go tools studio cfg copyenv --application cde2e951-d40a-3548-8b45-64c0ded97940 \
		--microservice-name frontend \
		--from-env test \
		--to-env prod
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logContext := logrus.StandardLogger()

		sourceMicroserviceName := viper.GetString("tools.studio.cfg.microservice-name")
		sourceEnvironment := viper.GetString("tools.studio.cfg.from-env")
		destinationEnvironment := viper.GetString("tools.studio.cfg.to-env")
		application := viper.GetString("tools.studio.cfg.application")

		sourceConfigMapName := fmt.Sprintf("%s-%s-env-variables", sourceEnvironment, sourceMicroserviceName)
		namespace := fmt.Sprintf("application-%s", application)
		logContext.WithFields(logrus.Fields{
			"k8sNamespace":     namespace,
			"k8sConfigmap":     sourceConfigMapName,
			"microserviceName": sourceMicroserviceName,
		}).Info("Begins to copy env var configmap")

		kubeCtlGetCfgMap := exec.Command("kubectl", "-n", namespace, "get", "configmap", sourceConfigMapName, "-o", "json")
		_, err := kubeCtlGetCfgMap.StderrPipe()
		if err != nil {
			logContext.Fatal(err)
			return
		}

		configMapJson, err := kubeCtlGetCfgMap.Output()
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"k8sNamespace": namespace,
				"k8sConfigMap": sourceConfigMapName,
				"error":        err,
			}).Fatal("Failed to get configmap")
			return
		}

		var modifiedConfigMap map[string]interface{}
		json.Unmarshal(configMapJson, &modifiedConfigMap)
		destinationConfigMap := fmt.Sprintf("%s-%s-env-variables", destinationEnvironment, sourceMicroserviceName)
		modifiedConfigMap["metadata"].(map[string]interface{})["name"] = destinationConfigMap
		// delete uid and version to ensure it patches
		delete(modifiedConfigMap["metadata"].(map[string]interface{}), "resourceVersion")
		delete(modifiedConfigMap["metadata"].(map[string]interface{}), "uid")
		out, _ := json.Marshal(modifiedConfigMap)

		kubectlApply := exec.Command("kubectl", "-o", "json", "apply", "-f", "-")
		kubectlApply.Stdin = bytes.NewReader(out)
		kubectlApply.Stderr = os.Stderr

		kubectlApply.Start()
		kubeCtlGetCfgMap.Run()
		kubectlApply.Wait()

		logContext.WithFields(logrus.Fields{
			"k8sNamespace":            namespace,
			"k8sSourceCondfigMap":     sourceConfigMapName,
			"k8sDestinationConfigMap": destinationConfigMap,
		}).Info("Successfully copied configmap for env variables")
	},
}
