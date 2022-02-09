package k8s

import (
	"errors"
	"fmt"
	"strings"

	dolittleK8s "github.com/dolittle/platform-api/pkg/dolittle/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateApplicationResource(config CreateResourceConfig, customerID string, application dolittleK8s.ShortInfo) *batchv1.Job {
	namespace := config.Namespace
	gitUserName := config.GitUserName
	gitUserEmail := config.GitUserEmail
	apiSecrets := config.ApiSecrets
	localBranch := config.LocalBranch
	remoteBranch := config.RemoteBranch
	serviceAccountName := config.ServiceAccountName
	platformImage := config.PlatformImage
	platformEnvironment := config.PlatformEnvironment
	isProduction := config.IsProduction

	terraformCustomerName := strings.ToLower(
		fmt.Sprintf("customer_%s", customerID),
	)

	applicationID := application.ID
	applicationName := application.Name

	terrformFileName := fmt.Sprintf("customer_%s_%s", customerID, applicationID)

	// Unique identifier of the job
	name := fmt.Sprintf("create-application-%s", applicationID)
	if len(name) >= 64 {
		panic("Not allowed due to kuberentes restriction")
	}
	annotations := platformK8s.GetAnnotationsForApplication(customerID, applicationID)

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					RestartPolicy:      "Never",
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
						gitSetup(platformImage, localBranch, gitUserEmail, gitUserName),
						// Create terraform
						// TODO We don't really need envfrom here
						createTerraform(platformImage, []string{
							"sh",
							"-c",
							fmt.Sprintf(`
/app/bin/app tools terraform create-customer-application-hcl \
--application-name="%s" \
--application-id="%s" \
--customer="%s"  > /pod-data/git/Source/V3/Azure/%s.tf;
`,
								applicationName,
								applicationID,
								terraformCustomerName,
								terrformFileName,
							),
						}),
						gitUpdateTerraform(platformImage, terrformFileName, localBranch, remoteBranch),
						// Update git with the changes

						// Terraform init new module
						terraformInit(platformImage),
						// Terraform apply new module
						terraformApply(platformImage, terrformFileName),
						// We are not outputing the applicaiton terraform
						terraformOutputJSON(platformImage),
						// TODO build-terraform-info
						// TODO make it so build-terraform-info works with 1 output not all
						toolsStudioBuildTerraformInfo(platformImage, platformEnvironment, customerID),
						gitUpdateStudioTerraformInfo(platformImage, platformEnvironment, customerID, localBranch, remoteBranch),

						{
							Name:            "build-application-in-cluster",
							ImagePullPolicy: "Always",
							Image:           platformImage,
							// TODO do we let it handle its own git?
							Env: []corev1.EnvVar{
								{
									Name:  "GIT_REPO_BRANCH", // Not really needed, but it fails without it
									Value: "na",
								},
								// Not needed as we are using GIT_REPO_DIRECTORY_ONLY
								{
									Name:  "GIT_REPO_URL", // Not needed
									Value: "na",
								},
								// Not needed as we are using GIT_REPO_DIRECTORY_ONLY
								{
									Name:  "GIT_REPO_SSH_KEY",
									Value: "na",
								},
								{
									Name:  "GIT_REPO_DIRECTORY",
									Value: "/pod-data/git/",
								},
								{
									Name:  "GIT_REPO_DIRECTORY_ONLY",
									Value: "true",
								},
								{
									Name:  "GIT_REPO_DRY_RUN",
									Value: "true",
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "platform-terraform-env-variables",
										},
									},
								},
							},
							Command: []string{
								"sh",
								"-c",
								fmt.Sprintf(
									`
/app/bin/app tools automate create-application \
--platform-environment="%s" \
--with-environments \
--with-welcome-microservice \
--customer-id="%s" \
--application-id="%s" \
--is-production="%t"
`,
									platformEnvironment,
									customerID,
									applicationID,
									isProduction,
								),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shared-data",
									MountPath: "/pod-data",
								},
								{
									Name:      "secrets",
									MountPath: "/dolittle/.ssh/operations",
									SubPath:   "operations",
								},
								{
									Name:      "secrets",
									MountPath: "/dolittle/.ssh/operations.pub",
									SubPath:   "operations.pub",
								},
							},
						},
						gitUpdate(platformImage, "post-application-created", []string{
							"sh",
							"-c",
							fmt.Sprintf(`
cd /pod-data/git;
git add ./Source/V3/platform-api/%s/%s;
git status;
git commit -m "Application created %s";
git log -1;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git push origin %s:%s;
							`,
								platformEnvironment,
								customerID,
								customerID,
								localBranch,
								remoteBranch,
							),
						}),
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

func DeleteApplicationResource() error {
	// TODO
	return errors.New("TODO: currently we lock this in azure")
}
