package requesthandler

import (
	_ "embed"
	"errors"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/sirupsen/logrus"

	. "github.com/dolittle-entropy/platform-api/pkg/platform/microservice/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/simple"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
)

type simpleHandler struct {
	*handler
	repo simple.Repo
}

func NewSimpleHandler(parser Parser, repo simple.Repo, gitRepo storage.Repo, logContext logrus.FieldLogger) Handler {
	return &simpleHandler{
		repo:    repo,
		handler: newHandler(parser, gitRepo, platform.MicroserviceKindBusinessMomentsAdaptor, logContext),
	}
}
func (s *simpleHandler) Create(inputBytes []byte, applicationInfo platform.Application) (platform.Microservice, *Error) {
	// Function assumes access check has taken place

	var ms platform.HttpInputSimpleInfo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		return nil, statusErr
	}

	ingress := CreateTodoIngress()

	// TODO I cant decide if domainNamePrefix or SecretNamePrefix is better
	//if ms.Extra.Ingress.SecretNamePrefix == "" {
	//	utils.RespondWithError(w, http.StatusBadRequest, "Missing extra.ingress.secretNamePrefix")
	//	return
	//}

	if ms.Extra.Ingress.Host == "" {
		return nil, NewBadRequest(errors.New("Missing extra.ingress.host"))
	}

	tenant, err := getFirstTenant(s.gitRepo, applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		return nil, NewInternalError(err)
	}

	err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, tenant, ms)
	if err != nil {
		return nil, NewInternalError(err)
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
		// TODO change
		return nil, NewInternalError(err)
	}

	return ms, nil
}

func (s *simpleHandler) Delete(namespace string, microserviceID string) *Error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return NewInternalError(err)
	}
	return nil
}
