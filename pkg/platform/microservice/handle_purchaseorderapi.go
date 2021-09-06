package microservice

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func (s *service) handlePurchaseOrderAPI(w http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place
	var ms platform.HttpInputPurchaseOrderInfo
	readMicroservice(&ms, inputBytes, applicationInfo)
	err = s.purchaseOrderAPIRepo.Create(namespace, tenant, application, ms)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.gitRepo.SaveMicroservice(
		ms.Dolittle.TenantID,
		ms.Dolittle.ApplicationID,
		ms.Environment,
		ms.Dolittle.MicroserviceID,
		ms,
	)

	if err != nil {
		// TODO change
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, ms)
}
