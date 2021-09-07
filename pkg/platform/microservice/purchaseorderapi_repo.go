package microservice

import (
	"context"
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type purchaseOrderAPIRepo struct {
	k8sClient      *kubernetes.Clientset
	rawDataLogRepo rawdatalog.RawDataLogIngestorRepo
	kind           platform.MicroserviceKind
}

// NewPurchaseOrderAPIRepo creates a new instance of purchaseorderapiRepo.
func NewPurchaseOrderAPIRepo(k8sClient *kubernetes.Clientset, rawDataLogRepo rawdatalog.RawDataLogIngestorRepo) purchaseOrderAPIRepo {
	return purchaseOrderAPIRepo{
		k8sClient,
		rawDataLogRepo,
		platform.MicroserviceKindPurchaseOrderAPI,
	}
}

// Create creates a new PurchaseOrderAPI microservice, and a RawDataLog and WebhookListener if they don't exist.
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

	ctx := context.TODO()

	if err := r.createRawDataLogIfNotExists(); err != nil {
		return err
	}

	if err := r.createWebhookListenerIfNotExists(); err != nil {
		return err
	}

	if err := r.createPurchaseOrderMicroservice(namespace, headImage, runtimeImage, microservice, ctx); err != nil {
		return err
	}

	if err := r.addWebhookEndpoints(); err != nil {
		return err
	}

	//TODO: Customise the config to adhere to how we create purchase order api
	// TODO: Add webhooks

	// TODO: add rawDataLogMicroserviceID
	// configFiles.Data = map[string]string{}
	// We store the config data into the config-Files for the service to pick up on
	// b, _ := json.MarshalIndent(input, "", "  ")
	// configFiles.Data["microservice_data_from_studio.json"] = string(b)
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

	return nil
}

// Delete stops the running purchase order api and deletes the kubernetes resources.
func (r purchaseOrderAPIRepo) Delete(namespace string, microserviceID string) error {
	ctx := context.TODO()

	deployment, err := r.stopDeployment(ctx, namespace, microserviceID)
	if err != nil {
		return err
	}

	listOpts := metaV1.ListOptions{
		LabelSelector: labels.FormatLabels(deployment.GetObjectMeta().GetLabels()),
	}

	if err = k8sDeleteConfigmaps(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteSecrets(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteServices(r.k8sClient, ctx, namespace, listOpts); err != nil {
		return err
	}

	if err = k8sDeleteDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return err
	}
	return nil
}

func (r purchaseOrderAPIRepo) stopDeployment(ctx context.Context, namespace, microserviceID string) (v1.Deployment, error) {
	deployment, err := k8sGetDeployment(r.k8sClient, ctx, namespace, microserviceID)
	if err != nil {
		return deployment, err
	}

	if err = k8sStopDeployment(r.k8sClient, ctx, namespace, &deployment); err != nil {
		return deployment, err
	}
	return deployment, nil
}

func (r purchaseOrderAPIRepo) createRawDataLogIfNotExists() error {
	return errors.New("Not implemented")
}

func (r purchaseOrderAPIRepo) createPurchaseOrderMicroservice(namespace, headImage, runtimeImage string, microservice k8s.Microservice, ctx context.Context) error {
	opts := metaV1.CreateOptions{}

	microserviceConfigmap := k8s.NewMicroserviceConfigmap(microservice, todoCustomersTenantID)
	deployment := k8s.NewDeployment(microservice, headImage, runtimeImage)
	service := k8s.NewService(microservice)
	configEnvVariables := k8s.NewEnvVariablesConfigmap(microservice)
	configFiles := k8s.NewConfigFilesConfigmap(microservice)
	configSecrets := k8s.NewEnvVariablesSecret(microservice)

	// ConfigMaps
	_, err := r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, microserviceConfigmap, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("microservice config map") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configEnvVariables, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config env variables") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configFiles, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config files") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Secrets(namespace).Create(ctx, configSecrets, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("config secrets") }) != nil {
		return err
	}
	_, err = r.k8sClient.CoreV1().Services(namespace).Create(ctx, service, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("service") }) != nil {
		return err
	}
	_, err = r.k8sClient.AppsV1().Deployments(namespace).Create(ctx, deployment, opts)
	if k8sHandleResourceCreationError(err, func() { k8sPrintAlreadyExists("deployment") }) != nil {
		return err
	}

	return nil
}

func (r purchaseOrderAPIRepo) createWebhookListenerIfNotExists() error {
	return errors.New("Not implemented")
}

func (r purchaseOrderAPIRepo) addWebhookEndpoints() error {
	return errors.New("Not implemented")
}

/*
---
apiVersion: v1
kind: Secret
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api-secret-env-variables
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4
type: Opaque
data:

---
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api-env-variables
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4
data:
  LOG_LEVEL: debug
  DATABASE_READMODELS_URL: "mongodb://dev-mongo.application-1649ad53-5200-4a42-bcd5-7d559e0eefd4.svc.cluster.local"
  DATABASE_READMODELS_NAME: "supplier_portal_dev_poapi_readmodels"
  NODE_ENV: "development"
  TENANT: "4d9bb23e-aa88-4be2-9dbd-f0bb3932e9d6"
  SERVER_PORT: "8080"
  NATS_CLUSTER_URL: "dev-rawdatalogv1-nats.application-1649ad53-5200-4a42-bcd5-7d559e0eefd4.svc.cluster.local:4222"
  NATS_START_FROM_BEGINNING: "false"
  LOG_OUTPUT_FORMAT: json

---
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api-config-files
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4
data:

*/

/*
---
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api-dolittle
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4

data:
  resources.json: |
    {
      "4d9bb23e-aa88-4be2-9dbd-f0bb3932e9d6": {
        "readModels": {
          "host": "mongodb://dev-mongo.application-1649ad53-5200-4a42-bcd5-7d559e0eefd4.svc.cluster.local:27017",
          "database": "supplier_portal_dev_poapi_readmodels",
          "useSSL": false
        },
        "eventStore": {
          "servers": [
            "dev-mongo.application-1649ad53-5200-4a42-bcd5-7d559e0eefd4.svc.cluster.local"
          ],
          "database": "supplier_portal_dev_poapi_eventstore"
        }
      }
    }

  event-horizons.json: |
    [
    ]

  event-horizon-consents.json: |
    {
    }

  microservices.json: |
    {
    }

  metrics.json: |
    {
      "Port": 9700
    }

  clients.json: |
    {
      "public": {
        "host": "localhost",
        "port": 50052
      },
      "private": {
        "host": "localhost",
        "port": 50053
      }
    }

  endpoints.json: |
    {
      "public": {
        "port": 50052
      },
      "private": {
        "port": 50053
      }
    }

  appsettings.json: |
    {
      "Logging": {
        "IncludeScopes": false,
        "LogLevel": {
          "Default": "Debug",
          "System": "Information",
          "Microsoft": "Information"
        },
        "Console": {
          "IncludeScopes": true,
          "TimestampFormat": "[yyyy-MM-dd HH:mm:ss] "
        }
      }
    }

---
apiVersion: v1
kind: Service
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4
spec:
  selector:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  ports:
    - port: 8080
      targetPort: api
      name: api

---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
    dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
    dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
  labels:
    tenant: Flokk
    application: SupplierPortal
    environment: Dev
    microservice: PurchaseOrderAPI
  name: dev-purchase-order-api
  namespace: application-1649ad53-5200-4a42-bcd5-7d559e0eefd4
spec:
  selector:
    matchLabels:
      tenant: Flokk
      application: SupplierPortal
      environment: Dev
      microservice: PurchaseOrderAPI

  template:
    metadata:
      annotations:
        dolittle.io/tenant-id: 388c0cc7-24b2-46a7-8735-b583ce21e01b
        dolittle.io/application-id: 1649ad53-5200-4a42-bcd5-7d559e0eefd4
        dolittle.io/microservice-id: f2586326-1057-495b-a12b-7f45193b9402
      labels:
        tenant: Flokk
        application: SupplierPortal
        environment: Dev
        microservice: PurchaseOrderAPI
    spec:
      containers:
        - name: head
          image: dolittle/integrations-m3-purchaseorders:2.2.3
          ports:
            - containerPort: 8080
              name: api
          envFrom:
            - configMapRef:
                name: dev-purchase-order-api-env-variables
            - secretRef:
                name: dev-purchase-order-api-secret-env-variables
          volumeMounts:
            - mountPath: /app/data
              name: config-files

        - name: runtime
          image: dolittle/runtime:6.1.0
          ports:
            - containerPort: 50052
              name: runtime
            - containerPort: 9700
              name: runtime-metrics
          volumeMounts:
            - mountPath: /app/.dolittle/tenants.json
              subPath: tenants.json
              name: tenants-config
            - mountPath: /app/.dolittle/resources.json
              subPath: resources.json
              name: dolittle-config
            - mountPath: /app/.dolittle/endpoints.json
              subPath: endpoints.json
              name: dolittle-config
            - mountPath: /app/.dolittle/event-horizon-consents.json
              subPath: event-horizon-consents.json
              name: dolittle-config
            - mountPath: /app/.dolittle/microservices.json
              subPath: microservices.json
              name: dolittle-config
            - mountPath: /app/.dolittle/metrics.json
              subPath: metrics.json
              name: dolittle-config
            - mountPath: /app/appsettings.json
              subPath: appsettings.json
              name: dolittle-config
      volumes:
        - name: tenants-config
          configMap:
            name: dev-tenants
        - name: dolittle-config
          configMap:
            name: dev-purchase-order-api-dolittle
        - name: config-files
          configMap:
            name: dev-purchase-order-api-config-files

*/
