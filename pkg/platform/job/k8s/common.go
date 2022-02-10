package k8s

import (
	"fmt"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

type CreateResourceConfig struct {
	PlatformImage       string
	PlatformEnvironment string
	IsProduction        bool
	Namespace           string
	GitUserName         string
	GitUserEmail        string
	ApiSecrets          string
	LocalBranch         string
	RemoteBranch        string
	ServiceAccountName  string
}

func CreateResourceConfigFromViper(v *viper.Viper) CreateResourceConfig {
	return CreateResourceConfig{
		PlatformImage:       v.GetString("tools.jobs.image.operations"),
		PlatformEnvironment: v.GetString("tools.server.platformEnvironment"),
		IsProduction:        v.GetBool("tools.server.isProduction"),
		Namespace:           "system-api",
		GitUserName:         v.GetString("tools.jobs.git.user.name"),
		GitUserEmail:        v.GetString("tools.jobs.git.user.email"),
		ApiSecrets:          v.GetString("tools.jobs.secrets.name"),
		LocalBranch:         v.GetString("tools.jobs.git.branch.local"),
		RemoteBranch:        v.GetString("tools.jobs.git.branch.remote"),
		ServiceAccountName:  "system-api-manager",
	}
}

func sshSetup() corev1.Container {
	return corev1.Container{
		Name:  "ssh-setup",
		Image: "busybox",
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
			`cp -r /dolittle/.ssh /pod-data;
chmod 600 /pod-data/.ssh/operations;
chmod 600 /pod-data/.ssh/operations.pub;
ls -lah /pod-data`,
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
	}
}

func gitSetup(image string, localBranch string, gitUserEmail string, gitUserName string) corev1.Container {
	return corev1.Container{
		Name:            "git-setup",
		ImagePullPolicy: "Always",
		Image:           image,
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
			fmt.Sprintf(`
mkdir -p /pod-data/git;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git clone  --branch %s $GIT_REPO_URL /pod-data/git;
cd /pod-data/git;
git config user.email "%s";
git config user.name "%s";
`,
				localBranch,
				gitUserEmail,
				gitUserName,
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

func createTerraform(image string, command []string) corev1.Container {
	return corev1.Container{
		Name:            "create-terraform",
		ImagePullPolicy: "Always",
		Image:           image,
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "platform-terraform-env-variables",
					},
				},
			},
		},
		Command: command,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
}

// name is currently the filename without .tf suffix
func gitUpdateTerraform(image string, name string, localBranch string, remoteBranch string) corev1.Container {
	commands := []string{
		"sh",
		"-c",
		fmt.Sprintf(`
cd /pod-data/git;
git add ./Source/V3/Azure/%s.tf;
git status;
git commit -m "Adding %s";
git log -1;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git push origin %s:%s;
`,
			name,
			name,
			localBranch,
			remoteBranch,
		),
	}
	return gitUpdate(image, "terraform", commands)
}

func terraformInit(image string) corev1.Container {
	return corev1.Container{
		Name:            "terraform-init",
		ImagePullPolicy: "Always",
		Image:           image,
		Env: []corev1.EnvVar{
			{
				Name:  "TF_IN_AUTOMATION",
				Value: "1",
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
		WorkingDir: "/pod-data/git/Source/V3/Azure",
		Command: []string{
			"sh",
			"-c",
			"terraform init",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
}

// name is currently the filename without .tf suffix
func terraformApply(image string, name string) corev1.Container {
	return corev1.Container{
		Name:            "terraform-apply",
		ImagePullPolicy: "Always",
		Image:           image,
		Env: []corev1.EnvVar{
			{
				Name:  "TF_IN_AUTOMATION",
				Value: "1",
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
		WorkingDir: "/pod-data/git/Source/V3/Azure",
		Command: []string{
			"sh",
			"-c",
			fmt.Sprintf(
				`terraform apply -target="module.%s" -auto-approve -no-color`,
				name,
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

func terraformOutputJSON(image string) corev1.Container {
	return corev1.Container{
		Name:            "terraform-output-json",
		ImagePullPolicy: "Always",
		Image:           image,
		Env: []corev1.EnvVar{
			{
				Name:  "TF_IN_AUTOMATION",
				Value: "1",
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
		WorkingDir: "/pod-data/git/Source/V3/Azure",
		Command: []string{
			"sh",
			"-c",
			"terraform output -json > azure.json",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
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

func toolsStudioBuildTerraformInfo(platformImage string, platformEnvironment string, customerID string) corev1.Container {
	return corev1.Container{
		Name:            "tools-studio-build-terraform-info",
		ImagePullPolicy: "Always",
		Image:           platformImage,
		// Env is used over EnvFrom
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
			fmt.Sprintf(`
/app/bin/app tools studio build-terraform-info \
--customer-id="%s" \
--platform-environment="%s" \
/pod-data/git/Source/V3/Azure/azure.json
`,
				customerID,
				platformEnvironment,
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

func gitUpdateStudioTerraformInfo(platformImage string, platformEnvironment string, customerID string, localBranch string, remoteBranch string) corev1.Container {
	commands := []string{
		"sh",
		"-c",
		fmt.Sprintf(`
cd /pod-data/git;
git add ./Source/V3/platform-api/%s/%s;
git status;
git commit -m "Adding terraform json to studio for customer %s";
git log -1;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git push origin %s:%s;
`,
			platformEnvironment,
			customerID,
			customerID,
			localBranch,
			remoteBranch,
		),
	}
	return gitUpdate(platformImage, "studio-terraform", commands)
}

func toolsStudioBuildStudioInfo(platformImage string, platformEnvironment string, customerID string) corev1.Container {
	return corev1.Container{
		Name:            "tools-studio-build-studio-info",
		ImagePullPolicy: "Always",
		Image:           platformImage,
		// Env is used over EnvFrom
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
			fmt.Sprintf(`
/app/bin/app tools studio build-studio-info \
--platform-environment="%s" \
--disable-environments=false %s
`,
				platformEnvironment,
				customerID,
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

func gitUpdateStudioInfo(platformImage string, platformEnvironment string, customerID string, localBranch string, remoteBranch string) corev1.Container {
	commands := []string{
		"sh",
		"-c",
		fmt.Sprintf(`
cd /pod-data/git;
git add ./Source/V3/platform-api/%s/%s;
git status;
git commit -m "Adding studio json to studio for customer %s";
git log -1;
GIT_SSH_COMMAND="ssh -i /pod-data/.ssh/operations -o IdentitiesOnly=yes -o StrictHostKeyChecking=no" git push origin %s:%s;
`,
			platformEnvironment,
			customerID,
			customerID,
			localBranch,
			remoteBranch,
		),
	}

	return gitUpdate(platformImage, "studio-info", commands)
}

func gitUpdate(platformImage string, suffix string, commands []string) corev1.Container {
	return corev1.Container{
		Name:            fmt.Sprintf("git-update-%s", suffix),
		ImagePullPolicy: "Always",
		Image:           platformImage,
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "platform-terraform-env-variables",
					},
				},
			},
		},
		Command: commands,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "shared-data",
				MountPath: "/pod-data",
			},
		},
	}
}
