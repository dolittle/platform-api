package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	"github.com/thoas/go-funk"
)

func GetUsersFromKratos(url string) ([]KratosUser, error) {
	users := make([]KratosUser, 0)

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Get(url)
	if err != nil {
		return users, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return users, err
	}

	err = json.Unmarshal(body, &users)
	if err != nil {
		return users, err
	}

	return users, nil
}

func GetUserFromListByEmail(users []KratosUser, email string) (KratosUser, error) {
	found := funk.Find(users, func(kratosUser KratosUser) bool {
		return kratosUser.Traits.Email == email
	})

	if found == nil {
		return KratosUser{}, errors.New("not-found")
	}

	return found.(KratosUser), nil
}

func UserCustomersContains(user KratosUser, customerID string) bool {
	exists := funk.Contains(user.Traits.Tenants, func(currentCustomerID string) bool {
		return customerID == currentCustomerID
	})

	return exists
}

func PrintPutUser(url string, user KratosUser, customerID string) string {
	_template := `
# Run this command to update the user to have access to {{.CustomerID}}
curl -X PUT http://localhost:4434/identities/{{.UserID}} \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' --data @- <<EOF
{
  "schema_id": "default",
  "traits": {
    "email": "{{.Email}}",
    "tenants": {{.Customers}}
  }
}
EOF
`

	type templateData struct {
		UserID     string
		Email      string
		Customers  string
		CustomerID string
	}

	customers := make([]string, 0)
	if len(user.Traits.Tenants) > 0 {
		customers = user.Traits.Tenants
	}

	customerAsBytes, _ := json.MarshalIndent(customers, "", " ")
	data := templateData{
		UserID:     user.ID,
		Email:      user.Traits.Email,
		Customers:  string(customerAsBytes),
		CustomerID: customerID,
	}

	// Create a new template and parse the letter into it.
	t := template.Must(template.New("").Parse(_template))
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		fmt.Println(err)
		return ""
	}

	return tpl.String()
}
