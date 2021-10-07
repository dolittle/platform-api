package purchaseorderapi

type service struct {
	purchaseOrderHandler *purchaseorderapi.Handler
	k8sDolittleRepo      platform.K8sRepo
	gitRepo              storage.Repo
	parser               parser.Parser
	logger               logrus.FieldLogger
}
