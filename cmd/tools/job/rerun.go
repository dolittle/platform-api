package job

import (
	"context"
	"fmt"
	"os"
	"time"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var rerunCMD = &cobra.Command{
	Use:   "rerun",
	Short: "Rerun a job",
	Long: `
Deletes and recreates a job to force it to rerun.

go run main.go tools job rerun <job-name>
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logger := logrus.StandardLogger()
		logContext := logger.WithFields(logrus.Fields{
			"command": "rerun",
		})

		jobName := args[0]
		if jobName == "" {
			logContext.Fatal("Can't rerun a job if we don't know what to look for, please add the job you want to rerun")
		}
		logContext = logContext.WithField("job_name", jobName)

		client, _ := platformK8s.InitKubernetesClient()

		ctx := context.TODO()
		namespace := "system-api"

		job, err := client.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			logContext.WithField("error", err.Error()).Fatal("failed to get the job")
		}

		watchOptions := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		watcher, err := client.BatchV1().Jobs(namespace).Watch(timeoutCtx, watchOptions)
		defer cancel()
		defer watcher.Stop()

		if err != nil {
			logContext.WithField("error", err.Error()).Fatal("failed to create the watcher for job deletion")
		}

		logContext.Info("deleting the job")
		err = client.BatchV1().Jobs(namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
		if err != nil {
			logContext.WithField("error", err.Error()).Fatal("failed to delete the job")
		}

		for {
			select {
			case event := <-watcher.ResultChan():
				if event.Type == watch.Deleted {
					logContext.Info("The job was deleted, recreating it")
					// have to cleanup some auto-generated properties from the fetched job
					// otherwise k8s will complain during creation
					delete(job.Spec.Template.ObjectMeta.Labels, "controller-uid")
					delete(job.Spec.Selector.MatchLabels, "controller-uid")
					job.ResourceVersion = ""

					_, err = client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
					if err != nil {
						logContext.WithField("error", err.Error()).Fatal("failed to create the job")
					}

					logContext.Info("done")
					return
				}

			case <-timeoutCtx.Done():
				logContext.Info("exiting, context timeout")
				return
			}
		}
	},
}
