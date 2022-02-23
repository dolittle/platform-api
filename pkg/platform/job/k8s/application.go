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
	gitRemote := config.GitRemote
	gitUserName := config.GitUserName
	gitUserEmail := config.GitUserEmail
	apiSecrets := config.ApiSecrets
	branch := config.GitBranch
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

	terraformBaseContainer := terraformBase(platformImage, apiSecrets)
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
						// We could write the env variables required?
						gitSetup(platformImage, gitRemote, branch, gitUserEmail, gitUserName),
						// Create terraform
						// TODO We don't really need envfrom here
						createTerraformWithCommand(terraformBaseContainer, []string{
							"sh",
							"-c",
							fmt.Sprintf(`
/app/bin/app tools terraform template application \
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

						gitUpdateTerraform(platformImage, terrformFileName, branch),
						// Update git with the changes

						// Terraform init new module
						terraformInit(terraformBaseContainer),
						// Terraform apply new module
						terraformApply(terraformBaseContainer, terrformFileName),
						// Terraform create azure.json
						terraformOutputJSON(terraformBaseContainer),
						toolsStudioBuildTerraformInfo(platformImage, platformEnvironment, customerID),
						gitUpdateStudioTerraformInfo(platformImage, platformEnvironment, customerID, branch),

						buildApplicationInCluster(platformImage, platformEnvironment, customerID, applicationID, isProduction),
						gitUpdate(platformImage, "post-application-created", []string{
							"sh",
							"-c",
							fmt.Sprintf(`
cd /pod-data/git;
git add ./Source/V3/platform-api/%s/%s;
git status;
git commit -m "Application created %s";
git log -1;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git push origin %s;
							`,
								platformEnvironment,
								customerID,
								customerID,
								branch,
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

// buildApplicationInCluster
// We rely on  next steps to write to git
func buildApplicationInCluster(platformImage string, platformEnvironment string, customerID string, applicationID string, isProduction bool) corev1.Container {

	envVars := []corev1.EnvVar{
		{
			Name:  "KUBECONFIG",
			Value: "incluster",
		},
	}
	envVars = append(envVars, envVarGitNotInUse()...)

	return corev1.Container{
		Name:            "build-application-in-cluster",
		ImagePullPolicy: "Always",
		Image:           platformImage,
		Env:             envVars,
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
		},
	}
}
