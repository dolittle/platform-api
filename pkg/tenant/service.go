package tenant

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func NewService() service {
	return service{}
}

func (s *service) Create(w http.ResponseWriter, r *http.Request) {
	var input HttpInput
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = json.Unmarshal(b, &input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	utils.RespondWithJSON(w, http.StatusOK, input)
}
