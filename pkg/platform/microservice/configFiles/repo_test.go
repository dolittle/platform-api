package configFiles_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	"github.com/dolittle/platform-api/pkg/platform/microservice/configFiles"
)

var _ = Describe("Repo", func() {

	var (
		repo configFiles.ConfigFilesRepo

		clientSet *fake.Clientset
		config    *rest.Config
		k8sRepo   platformK8s.K8sRepo
		// k8sRepoV2 k8s.Repo
		logger         *logrus.Logger
		applicationID  string
		environment    string
		microserviceID string
	)

	BeforeEach(func() {

		clientSet = &fake.Clientset{}
		config = &rest.Config{}
		logger, _ = logrusTest.NewNullLogger()
		k8sRepo = platformK8s.NewK8sRepo(clientSet, config, logger.WithField("context", "k8s-repo"))
		// k8sRepoV2 = k8s.NewRepo(clientSet, logger.WithField("context", "k8s-repo-v2"))
		repo = configFiles.NewConfigFilesK8sRepo(k8sRepo, clientSet, logger)
		applicationID = "53fd6176-4a6b-4bb0-a32d-b18c69607e78"
		environment = "test"
		microserviceID = "963a5d70-8652-494a-a843-d3bebd660acb"

		// have to mock this so that we can mock GetDeploymentsByEnvironmentWithMicroservice() from pkg/k8s/repo
		// TODO have the platformK8s.NewK8sRepo take the k8sRepoV2 interface(s) as an argument so that we could mock it
		// instead of having to go this deep in the chain
		clientSet.AddReactor("list", "deployments", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
			filters := action.(testing.ListActionImpl).ListRestrictions
			Expect(filters.Labels.Matches(labels.Set{
				"tenant":       string(selection.Exists),
				"environment":  environment,
				"microservice": string(selection.Exists),
				"application":  string(selection.Exists),
			})).To(BeTrue())

			deployment := &appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-deployment",
							Labels: map[string]string{
								"tenant":       "fake-tenant",
								"application":  "fake-application",
								"environment":  environment,
								"microservice": "fake-microservice",
							},
							Annotations: map[string]string{
								"dolittle.io/microservice-id": microserviceID,
								"dolittle.io/application-id":  applicationID,
							},
						},
					},
				},
			}

			return true, deployment, nil
		})
	})

	Describe("Adding a new entry", func() {
		When("the incoming data is utf8", func() {
			It("should save it to 'data' property", func() {
				configFile := configFiles.MicroserviceConfigFile{
					Name:  "utf8-valid-file.txt",
					Value: []byte("𝝸м ᶏ ṕξꝛғȩçƭɬý 𝘀ɑᵮҿ ựⱦ𝐟8 ŝᴛгï𝖞𝓰 𝙣ө ᴨḗ𝑒ḑ 𝓽ⲟ ŵ𝞸ɾṙỳ ẃḩỵ àᵳḙ γ੦ǜ 𝓇𝕦ṇ𝜋𝙞𝖞ǵ 𝒸𝜎мξ ƀ⍺ć𝞳 𝛊 ѡợոṯ ḫṻɼ𝔱 ɏ𝗈ṵ"),
				}

				clientSet.AddReactor("update", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					updateAction := action.(testing.UpdateAction)
					originalObj := updateAction.GetObject()
					configMap := originalObj.(*corev1.ConfigMap)

					Expect(configMap.Data[configFile.Name]).To(Equal(string(configFile.Value)))
					Expect(configMap.BinaryData).To(BeEmpty())

					return true, configMap, nil
				})

				err := repo.AddEntryToConfigFiles(applicationID, environment, microserviceID, configFile)

				Expect(err).To(BeNil())
			})
		})

		When("the incoming data is not utf8", func() {
			It("should save it to 'binaryData' property", func() {
				configFile := configFiles.MicroserviceConfigFile{
					Name:  "not-utf8-valid-file.txt",
					Value: []byte{0xff, 0xfe, 0xfd},
				}

				clientSet.AddReactor("update", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					updateAction := action.(testing.UpdateAction)
					originalObj := updateAction.GetObject()
					configMap := originalObj.(*corev1.ConfigMap)

					Expect(configMap.BinaryData[configFile.Name]).To(Equal(configFile.Value))
					Expect(configMap.Data).To(BeEmpty())

					return true, configMap, nil
				})

				err := repo.AddEntryToConfigFiles(applicationID, environment, microserviceID, configFile)

				Expect(err).To(BeNil())
			})
		})

	})

	Describe("getting config files names", func() {
		When("the config map has both data and binary data", func() {
			It("should return them both", func() {
				binaryFile := "not-a-virus.mp4.mp3.zip.mov.exe"
				dataFile := "just-data.txt"

				clientSet.AddReactor("get", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					getAction := action.(testing.GetAction)
					configMap := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name: getAction.GetName(),
						},
						BinaryData: map[string][]byte{
							binaryFile: {0xff, 0xfe, 0xfd},
						},
						Data: map[string]string{
							dataFile: "i just be data",
						},
					}
					return true, configMap, nil
				})

				files, err := repo.GetConfigFilesNamesList(applicationID, environment, microserviceID)
				Expect(err).To(BeNil())
				Expect(files).ToNot(BeEmpty())
				Expect(files).To(ConsistOf(binaryFile, dataFile))
			})
		})
	})

	Describe("deleting an entry from config files", func() {
		It("should delete the entry from the binaryData property", func() {
			fileToDelete := "binary-file"
			fileToKeep := "second-binary-file"

			clientSet.AddReactor("get", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				getAction := action.(testing.GetAction)
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: getAction.GetName(),
					},
					BinaryData: map[string][]byte{
						fileToDelete: {0xfe, 0xfe, 0xfd},
						fileToKeep:   {0xff, 0xff, 0xff},
					},
				}
				return true, configMap, nil
			})

			clientSet.AddReactor("update", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				updateAction := action.(testing.UpdateAction)
				originalObj := updateAction.GetObject()
				configMap := originalObj.(*corev1.ConfigMap)

				Expect(configMap.BinaryData[fileToKeep]).ToNot(BeNil())
				Expect(configMap.BinaryData[fileToDelete]).To(BeNil())

				return true, configMap, nil
			})

			err := repo.RemoveEntryFromConfigFiles(applicationID, environment, microserviceID, fileToDelete)
			Expect(err).To(BeNil())
		})
		It("should delete the entry from the data property", func() {
			fileToDelete := "data-file"
			fileToKeep := "second-data-file"

			clientSet.AddReactor("get", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				getAction := action.(testing.GetAction)
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: getAction.GetName(),
					},
					Data: map[string]string{
						fileToDelete: "im going to be deleted",
						fileToKeep:   "hähä i won't >:D",
					},
				}

				return true, configMap, nil
			})

			clientSet.AddReactor("update", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				updateAction := action.(testing.UpdateAction)
				originalObj := updateAction.GetObject()
				configMap := originalObj.(*corev1.ConfigMap)

				Expect(configMap.Data[fileToKeep]).ToNot(BeEmpty())
				Expect(configMap.Data[fileToDelete]).To(BeEmpty())

				return true, configMap, nil
			})

			err := repo.RemoveEntryFromConfigFiles(applicationID, environment, microserviceID, fileToDelete)

			Expect(err).To(BeNil())
		})

	})
})
