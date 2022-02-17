package studio

import (
	"fmt"
	"os"

	"github.com/dolittle/platform-api/pkg/k8s"
	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	rbacv1 "k8s.io/api/rbac/v1"
)

var userAdminCMD = &cobra.Command{
	Use:   "admin [user] [add|remove|list] [useruuid]",
	Short: "Add or remove a user from the studio admin access",
	Long: `


List users
	go run main.go tools studio admin user list

Add user "fake"
	go run main.go tools studio admin user add fake

Remove user "fake"
	go run main.go tools studio admin user remove fake
	`,
	Args: cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		logger := logrus.StandardLogger()

		logContext := logger.WithField("cmd", "build-studio-info")

		what := args[0]
		if what != "user" {
			fmt.Println("Only support user today")
			return
		}

		do := args[1]
		if !funk.Contains([]string{"add", "remove", "list"}, do) {
			fmt.Println("Only allowed to add, remove or list")
			return
		}

		userUUID := ""
		if len(args) == 3 {
			userUUID = args[2]
		}

		k8sClient, _ := platformK8s.InitKubernetesClient()
		k8sRepoV2 := k8s.NewRepo(k8sClient, logger.WithField("context", "k8s-repo-v2"))

		namespace := "system-api"
		name := "platform-admin"
		newSubject := rbacv1.Subject{
			Kind:     rbacv1.UserKind,
			APIGroup: "rbac.authorization.k8s.io",
			Name:     userUUID,
		}

		var err error
		switch do {
		case "add":
			err = k8sRepoV2.AddSubjectToRoleBinding(namespace, name, newSubject)
		case "remove":
			err = k8sRepoV2.RemoveSubjectToRoleBinding(namespace, name, newSubject)
		case "list":
			rolebinding, err := k8sRepoV2.GetRoleBinding(namespace, name)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Error("Getting")
				return
			}

			userIDS := funk.Map(rolebinding.Subjects, func(subject rbacv1.Subject) string {
				return subject.Name
			})
			logContext.WithFields(logrus.Fields{
				"users": userIDS,
			}).Info("users with access")
		default:
			// Shouldn't happen
			fmt.Println("Unsupported do")
			return
		}

		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("Saving")
			return
		}

		logContext.Info("Done!")
	},
}
