package testutils

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func CheckJSONPrettyPrint(a string, b string) {
	raw := []byte(b)
	var rawI interface{}
	json.Unmarshal(raw, &rawI)
	jsonBytes, _ := json.MarshalIndent(rawI, "", "  ")
	expect := string(jsonBytes)
	Expect(a).To(Equal(expect))
}
