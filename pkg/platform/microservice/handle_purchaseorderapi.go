package microservice

import (
	_ "embed"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/platform"
)

func (s *service) handlePurchaseOrderAPI(w http.ResponseWriter, r *http.Request, inputBytes []byte, applicationInfo platform.Application) {
	// Function assumes access check has taken place

}
