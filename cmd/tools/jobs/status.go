package jobs

import (
	"context"
	"fmt"

	platformK8s "github.com/dolittle/platform-api/pkg/platform/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var statusCMD = &cobra.Command{
	Use:   "status",
	Short: "Get status for a job",
	Long: `
	Outputs a State of each container in the job

	go run main.go tools jobs status XXX
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Hmm it is on the pod level
		// Init:Error, Completed, Init:5/9, Init:ErrImagePull, Init:ImagePullBackOff
		jobID := args[0]
		if jobID == "" {
			fmt.Println("We cant give you the status of a job if we don't know what to look for, please add the job you want the status for")
			return
		}

		client, _ := platformK8s.InitKubernetesClient()

		ctx := context.TODO()
		namespace := "system-api"

		resource, err := client.BatchV1().Jobs(namespace).Get(ctx, jobID, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, condition := range resource.Status.Conditions {
			fmt.Println(condition.Type)
		}

		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", jobID),
		})

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, pod := range pods.Items {

			fmt.Println(pod.Status.Phase, pod.Status.StartTime)
			for _, status := range pod.Status.InitContainerStatuses {
				cmd := fmt.Sprintf(`kubectl -n %s logs %s -c %s`,
					pod.Namespace,
					pod.Name,
					status.Name,
				)
				fmt.Println(cmd)

				if status.State.Running != nil {
					fmt.Println("Running", status.State.Running.StartedAt)
				}

				if status.State.Waiting != nil {
					fmt.Println("Waiting", status.State.Waiting.Message, status.State.Waiting.Reason)
				}

				if status.State.Terminated != nil {
					fmt.Println("Terminated", status.State.Terminated.Message, status.State.Terminated.Reason)
				}

				fmt.Println("")
				// req := client.CoreV1().Pods(namespace).GetLogs(podName, &podLogOptions)
			}
		}
	},
}

func init() {

}
