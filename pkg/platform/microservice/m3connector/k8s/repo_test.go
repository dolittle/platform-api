package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repo", func() {

	XIt("", func() {

		// getCalled := false
		// clientSet.AddReactor("get", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
		// 	getCalled = true
		// 	getAction := action.(testing.GetAction)
		// 	name := getAction.GetName()
		// 	expectedName := fmt.Sprintf("%s-kafka-files", environment)
		// 	Expect(name).To(Equal(expectedName))

		// 	return true, nil, k8sErrors.NewNotFound(schema.ParseGroupResource("corev1.configmaps"), name)
		// })
		// createCalled := false
		// clientSet.AddReactor("create", "configmaps", func(action testing.Action) (bool, runtime.Object, error) {
		// 	createCalled = true
		// 	createAction := action.(testing.CreateAction)
		// 	originalObj := createAction.GetObject()
		// 	configMap := originalObj.(*corev1.ConfigMap)

		// 	Expect(configMap.Data["accessKey.pem"]).To(Equal(""))

		// 	return true, configMap, nil
		// })

	})

})
