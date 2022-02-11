package platform

import "fmt"

func GetCustomerGroup(customerID string) string {
	return fmt.Sprintf("tenant-%s", customerID)
}
