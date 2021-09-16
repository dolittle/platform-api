package requesthandler

import (
	_ "embed"
	"errors"
	"net/http"
	"strings"

	"github.com/dolittle-entropy/platform-api/pkg/dolittle/k8s"
	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/platform/microservice/rawdatalog"
	"github.com/dolittle-entropy/platform-api/pkg/platform/storage"
	"github.com/thoas/go-funk"
)

type rawDataLogIngestorHandler struct {
	parser  Parser
	repo    rawdatalog.RawDataLogIngestorRepo
	gitRepo storage.Repo
}

func NewRawDataLogIngestorHandler(parser Parser, repo rawdatalog.RawDataLogIngestorRepo, gitRepo storage.Repo) Handler {
	return &rawDataLogIngestorHandler{parser, repo, gitRepo}
}
func (s *rawDataLogIngestorHandler) CanHandle(kind platform.MicroserviceKind) bool {
	return kind == platform.MicroserviceKindRawDataLogIngestor
}

func (s *rawDataLogIngestorHandler) Create(request *http.Request, inputBytes []byte, applicationInfo platform.Application) (platform.Microservice, *Error) {
	// Function assumes access check has taken place
	var ms platform.HttpInputRawDataLogIngestorInfo

	// TODO: Fix this when we get to work on rawDataLogIngestorRepo
	msK8sInfo, statusErr := s.parser.Parse(inputBytes, &ms, applicationInfo)
	if statusErr != nil {
		return nil, statusErr
	}
	storedIngress, err := getFirstIngress(s.gitRepo, applicationInfo.Tenant.ID, applicationInfo.ID, ms.Environment)
	if err != nil {
		return nil, NewInternalError(err)
	}
	ingress := k8s.Ingress{
		Host:       storedIngress.Host,
		SecretName: storedIngress.SecretName,
	}
	// TODO changing writeTo will break this.
	// TODO does this exist?
	if ms.Extra.WriteTo == "" {
		ms.Extra.WriteTo = "nats"
	}

	writeToCheck := funk.Contains([]string{
		"stdout",
		"nats",
	}, func(filter string) bool {
		return strings.HasSuffix(ms.Extra.WriteTo, filter)
	})

	if !writeToCheck {
		return nil, NewForbidden(errors.New("writeTo is not valid, leave empty or set to stdout"))
	}

	// TODO lookup to see if it exists?
	exists := false
	//exists := true s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//exists, err := s.rawDataLogIngestorRepo.Exists(namespace, ms.Environment, ms.Dolittle.MicroserviceID)
	//if err != nil {
	//	// TODO change
	//	utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	//	return
	//}
	if !exists {
		// Create in Kubernetes
		err = s.repo.Create(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms) //TODO:
	} else {
		err = s.repo.Update(msK8sInfo.Namespace, msK8sInfo.Tenant, msK8sInfo.Application, ingress, ms) //TODO:
	}

	if err != nil {
		// TODO change
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
		return nil, NewInternalError(err)
	}

	return ms, nil
}

func (s *rawDataLogIngestorHandler) Delete(namespace string, microserviceID string) *Error {
	if err := s.repo.Delete(namespace, microserviceID); err != nil {
		return NewInternalError(err)
	}
	return nil
}
