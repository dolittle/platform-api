package requesthandler

import (
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/businessmomentsadaptor"
	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
)

type businessMomentsAdapterHandler struct {
	*handler
	repo businessmomentsadaptor.Repo
}

func NewBusinessMomentsAdapterHandler(parser Parser, repo businessmomentsadaptor.Repo, gitRepo storage.Repo, logContext logrus.FieldLogger) Handler {
	return &businessMomentsAdapterHandler{
		repo:    repo,
		handler: newHandler(parser, gitRepo, platform.MicroserviceKindBusinessMomentsAdaptor, logContext),
	}
}

func (s *businessMomentsAdapterHandler) Create(inputBytes []byte, applicationInfo platform.Application) (platform.Microservice, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputBusinessMomentAdaptorInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		return ms, statusErr
	}
	ingress := CreateTodoIngress()

	tenant, err := getFirstTenant(s.gitRepo, applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		return nil, NewInternalError(err)
	}
	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, tenant, ms)
	if err != nil {
		// TODO change
		return ms, NewInternalError(err)
	}

	// TODO this could be an event
	// TODO this should be decoupled
	err = s.gitRepo.SaveMicroservice(
		ms.Dolittle.TenantID,
		ms.Dolittle.ApplicationID,
		ms.Environment,
		ms.Dolittle.MicroserviceID,
		ms,
	)
	if err != nil {
		return ms, NewInternalError(err)
	}
	return ms, nil
}

func (s *businessMomentsAdapterHandler) Delete(namespace string, microserviceID string) *Error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return NewInternalError(err)
	}
	return nil
}
