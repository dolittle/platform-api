package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// envVarForTerraform sets the ARM_* environment variables to their values
// from the Azure Service Principal that is configured inside the apiSecrets k8s secret.
// The ARM_* env variables are the default env vars that Terraform uses to setup
// it's access to Azure:
// https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret#configuring-the-service-principal-in-terraform
func envVarForTerraform(apiSecrets string) []corev1.EnvVar {
	gitVars := envVarGitNotInUse()
	terraformVars := []corev1.EnvVar{
		{
			Name: "ARM_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiSecrets,
					},
					Key: "ARM_CLIENT_ID",
				},
			},
		},
		{
			Name: "ARM_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiSecrets,
					},
					Key: "ARM_CLIENT_SECRET",
				},
			},
		},
		{
			Name: "ARM_SUBSCRIPTION_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiSecrets,
					},
					Key: "ARM_SUBSCRIPTION_ID",
				},
			},
		},
		{
			Name: "ARM_TENANT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiSecrets,
					},
					Key: "ARM_TENANT_ID",
				},
			},
		},
	}

	return append(gitVars, terraformVars...)
}

func createTerraformWithCommand(base corev1.Container, command []string) corev1.Container {
	copy := base
	copy.Name = "terraform-create"
	copy.Command = command
	return copy
}

func terraformBase(image string, apiSecrets string) corev1.Container {
	return corev1.Container{
		Name:            "terraform-base",
		ImagePullPolicy: "Always",
		Image:           image,
		Env:             envVarForTerraform(apiSecrets),
		WorkingDir:      "/pod-data/git/Source/V3/Azure",
		Command: []string{
			"sh",
			"-c",
			`echo "replace"`,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
}

func terraformInit(base corev1.Container) corev1.Container {
	copy := base
	copy.Name = "terraform-init"
	copy.Command = []string{
		"sh",
		"-c",
		"terraform init",
	}
	return copy
}

// name is currently the filename without .tf suffix
func terraformApply(base corev1.Container, name string) corev1.Container {
	copy := base
	copy.Name = "terraform-apply"
	copy.Command = []string{
		"sh",
		"-c",
		fmt.Sprintf(
			`terraform apply -target="module.%s" -auto-approve -no-color`,
			name,
		),
	}
	return copy
}

func terraformOutputJSON(base corev1.Container) corev1.Container {
	copy := base
	copy.Name = "terraform-output-json"
	copy.Command = []string{
		"sh",
		"-c",
		"terraform output -json > azure.json",
	}
	return copy
}

// terraformRemoveOutputJSON we remove the temp json to reduce the chance of data leaking
func terraformRemoveOutputJSON(image string) corev1.Container {
	return corev1.Container{
		Name:            "terraform-rm-output-json",
		ImagePullPolicy: "Always",
		Image:           image,
		WorkingDir:      "/pod-data/git/Source/V3/Azure",
		Command: []string{
			"sh",
			"-c",
			"rm azure.json",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
}
