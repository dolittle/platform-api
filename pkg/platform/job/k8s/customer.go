package k8s

import (
	"context"
	"errors"
	"fmt"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateCustomerResource(config CreateResourceConfig, customer dolittleK8s.ShortInfo) *batchv1.Job {
	namespace := config.Namespace
	gitRemote := config.GitRemote
	gitUserName := config.GitUserName
	gitUserEmail := config.GitUserEmail
	apiSecrets := config.ApiSecrets
	localBranch := config.LocalBranch
	remoteBranch := config.RemoteBranch
	platformImage := config.PlatformImage
	platformEnvironment := config.PlatformEnvironment
	// config.ServiceAccountName not in use
	// config.IsProduction not in use

	customerID := customer.ID
	customerName := customer.Name

	terrformFileName := fmt.Sprintf("customer_%s", customerID)

	// Unique identifier of the job
	name := fmt.Sprintf("create-customer-%s", customerID)

	labels := platformK8s.GetLabelsForCustomer(customerName)
	annotations := platformK8s.GetAnnotationsForCustomer(customerID)

	terraformBaseContainer := terraformBase(platformImage, apiSecrets)
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Volumes: []corev1.Volume{
						{
							Name:         "shared-data",
							VolumeSource: corev1.VolumeSource{},
						},
						{
							Name: "secrets",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: apiSecrets,
									Items: []corev1.KeyToPath{
										{
											Key:  "SSH_KEY_PUBLIC",
											Path: "operations.pub",
										},
										{
											Key:  "SSH_KEY_PRIVATE",
											Path: "operations",
										},
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						sshSetup(),
						gitSetup(platformImage, gitRemote, localBranch, gitUserEmail, gitUserName),
						createTerraformWithCommand(terraformBaseContainer, []string{
							"sh",
							"-c",
							fmt.Sprintf(`
/app/bin/app tools terraform create-customer-hcl \
--platform-environment="%s" \
--name="%s" \
--id="%s" \
> /pod-data/git/Source/V3/Azure/%s.tf;
`,
								platformEnvironment,
								customerName,
								customerID,
								terrformFileName,
							),
						}),
						gitUpdateTerraform(platformImage, terrformFileName, localBranch, remoteBranch),
						// Update git with the changes

						// Terraform init new module
						terraformInit(terraformBaseContainer),
						// Terraform apply new module
						terraformApply(terraformBaseContainer, terrformFileName),
						// Terraform create azure.json
						terraformOutputJSON(terraformBaseContainer),
						toolsStudioBuildTerraformInfo(platformImage, platformEnvironment, customerID),
						gitUpdateStudioTerraformInfo(platformImage, platformEnvironment, customerID, localBranch, remoteBranch),
						toolsStudioBuildStudioInfo(platformImage, platformEnvironment, customerID),
						gitUpdateStudioInfo(platformImage, platformEnvironment, customerID, localBranch, remoteBranch),
						terraformRemoveOutputJSON(platformImage),
					},
					Containers: []corev1.Container{
						{
							Name:  "summary",
							Image: "busybox",
							Command: []string{
								"sh",
								"-c",
								`echo "jobs done"`,
							},
						},
					},
				},
			},
		},
	}
}

func DoJob(client kubernetes.Interface, job *batchv1.Job) error {
	ctx := context.TODO()
	namespace := job.ObjectMeta.Namespace

	_, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		return nil
	}
	return nil
}

func DeleteCustomerResource() error {
	// TODO
	return errors.New("TODO: currently we lock this in azure")
}
