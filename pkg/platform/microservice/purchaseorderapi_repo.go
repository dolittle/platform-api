package microservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	v1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type purchaseOrderAPIRepo struct {
	client         *kubernetes.Clientset
	rawDataLogRepo rawdatalog.RawDataLogIngestorRepo
	kind           string
}

// Creates a new instance of purchaseorderapiRepo
func NewPurchaseOrderAPIRepo(k8sClient *kubernetes.Clientset, rawDataLogRepo rawdatalog.RawDataLogIngestorRepo) purchaseOrderAPIRepo {
	return purchaseOrderAPIRepo{
		k8sClient,
		rawDataLogRepo,
		platform.PurchaseOrderAPI,
	}
}

func (r purchaseOrderAPIRepo) Create(namespace string, tenant k8s.Tenant, application k8s.Application, input platform.HttpInputPurchaseOrderInfo) error {
	// TODO not sure where this comes from really, assume dynamic

	environment := input.Environment

	microserviceID := input.Dolittle.MicroserviceID
	microserviceName := input.Name
	headImage := input.Extra.Headimage
	runtimeImage := input.Extra.Runtimeimage

	microservice := k8s.Microservice{
		ID:          microserviceID,
		Name:        microserviceName,
		Tenant:      tenant,
		Application: application,
		Environment: environment,
		ResourceID:  todoCustomersTenantID,
		Kind:        r.kind,
	}

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, todoCustomersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)

	// TODO: Add webhooks
	// TODO: add rawDataLogMicroserviceID
	configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	b, _ := json.MarshalIndent(input, "", "  ")
	configFiles.Data["microservice_data_from_studio.json"] = string(b)

	// TODO lookup to see if it exists?
	// exists, err := r.rawDataLogRepo.Exists(namespace, environment, microserviceID)
	//exists, err := s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//if err != nil {
	//	// TODO change
	//	utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}

	// if !exists {
	// 	fmt.Println("Raw Data Log does not exist")
	// }

	// Assuming the namespace exists
	var err error
	ctx := context.TODO()

	opts := metaV1.CreateOptions{}
	// ConfigMaps
	_, err = r.client.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, opts)
	if handleResourceCreationError(err, "microserviceConfigMap") != nil {
		return err
	}
	_, err = r.client.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, opts)
	if handleResourceCreationError(err, "configEnvVariables") != nil {
		return err
	}
	_, err = r.client.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, opts)
	if handleResourceCreationError(err, "configFiles") != nil {
		return err
	}
	_, err = r.client.CoreV1().Secrets(namespace).Create(ctx, configSecrets, opts)
	if handleResourceCreationError(err, "configSecrets") != nil {
		return err
	}
	_, err = r.client.CoreV1().Services(namespace).Create(ctx, service, opts)
	if handleResourceCreationError(err, "service") != nil {
		return err
	}
	_, err = r.client.AppsV1().Deployments(namespace).Create(ctx, deployment, opts)
	if handleResourceCreationError(err, "deployment") != nil {
		return err
	}

	return nil
}

func (r purchaseOrderAPIRepo) Delete(namespace string, microserviceID string) error {
	// client := r.client
	// ctx := context.TODO()
	// // Not possible to filter based on annotations
	// deployment, err := r.getDeployment(ctx, namespace, microserviceID)
	// if err != nil {
	// 	return err
	// }

	// err = r.scaleDownDeployment(ctx, namespace, deployment)
	// if err != nil {
	// 	return err
	// }

	// // Selector information for microservice, based on labels
	// listOpts := metaV1.ListOptions{
	// 	LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	// }

	// // Remove configmaps
	// configs, _ := client.CoreV1().ConfigMaps(namespace).List(ctx, listOpts)

	// for _, config := range configs.Items {
	// 	err = client.CoreV1().ConfigMaps(namespace).Delete(ctx, config.Name, metaV1.DeleteOptions{})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return errors.New("todo")
	// 	}
	// }

	// // Remove secrets
	// secrets, _ := client.CoreV1().Secrets(namespace).List(ctx, listOpts)
	// for _, secret := range secrets.Items {
	// 	err = client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metaV1.DeleteOptions{})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return errors.New("todo")
	// 	}
	// }

	// // Remove Ingress
	// ingresses, _ := client.NetworkingV1().Ingresses(namespace).List(ctx, listOpts)
	// for _, ingress := range ingresses.Items {
	// 	err = client.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metaV1.DeleteOptions{})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return errors.New("issue")
	// 	}
	// }

	// // Remove Network Policy
	// policies, _ := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, listOpts)
	// for _, policy := range policies.Items {
	// 	err = client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policy.Name, metaV1.DeleteOptions{})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return errors.New("issue")
	// 	}
	// }

	// // Remove Service
	// services, _ := client.CoreV1().Services(namespace).List(ctx, listOpts)
	// for _, service := range services.Items {
	// 	err = client.CoreV1().Services(namespace).Delete(ctx, service.Name, metaV1.DeleteOptions{})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return errors.New("issue")
	// 	}
	// }

	// // Remove deployment
	// err = client.AppsV1().
	// 	Deployments(namespace).
	// 	Delete(ctx, deployment.Name, metaV1.DeleteOptions{})
	// if err != nil {
	// 	log.Fatal(err)
	// 	return errors.New("todo")
	// }

	return nil
}

func (r *purchaseOrderAPIRepo) getDeployment(context context.Context, namespace string, microserviceID string) (*v1.Deployment, error) {
	deployments, err := r.client.AppsV1().Deployments(namespace).List(context, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	found := false
	// Ugly name
	var foundDeployment v1.Deployment
	for _, deployment := range deployments.Items {
		_, ok := deployment.ObjectMeta.Labels["microservice"]
		if !ok {
			continue
		}

		if deployment.ObjectMeta.Annotations["dolittle.io/microservice-id"] == microserviceID {
			found = true
			foundDeployment = deployment
			break
		}
	}

	if !found {
		return nil, errors.New("not-found")
	}
	return &foundDeployment, nil
}

func (r *purchaseOrderAPIRepo) scaleDownDeployment(context context.Context, namespace string, deployment *v1.Deployment) error {
	s, err := r.client.AppsV1().
		Deployments(namespace).
		GetScale(context, deployment.Name, metaV1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return errors.New("issue")
	}

	sc := *s
	if sc.Spec.Replicas != 0 {
		sc.Spec.Replicas = 0
		_, err := r.client.AppsV1().
			Deployments(namespace).
			UpdateScale(context, deployment.Name, &sc, metaV1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
			return errors.New("todo")
		}
	}
	return nil
}

func handleResourceCreationError(creationError error, resourceName string) error {
	if creationError != nil {
		if !k8serrors.IsAlreadyExists(creationError) {
			log.Fatal(creationError)
			return errors.New("issue")
		}
		fmt.Printf("Skipping %s already exists\n", resourceName)
	}
	return nil
}
