package codegenerator

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type codegeneratorclient struct {
	baseUrl string
}

type M3ConnectorConsumerConfig struct {
	SolutionName string
	Environment  string
	Username     string
}

type KafkaConfig struct {
	BrokerURL            string
	InputTopic           string
	CommandTopic         string
	CommandReceiptsTopic string
	ChangeEventsTopic    string
	AccessKey            io.Reader
	Certificate          io.Reader
	Ca                   io.Reader
}

func newCodeGeneratorClient(baseUrl string) codegeneratorclient {
	return codegeneratorclient{
		baseUrl: baseUrl,
	}
}

func (c codegeneratorclient) GenerateM3ConnectorConsumer(zipFileName string,
	solutionName string,
	environment string,
	username string,
	kafkaConfig KafkaConfig) []byte {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{Transport: tr}

	query := url.Values{}
	query.Add("zipFileName", zipFileName)
	query.Add("solutionName", solutionName)
	query.Add("environment", environment)
	query.Add("username", username)
	query.Add("brokerUrl", kafkaConfig.BrokerURL)
	query.Add("inputTopic", kafkaConfig.InputTopic)
	query.Add("commandTopic", kafkaConfig.CommandTopic)
	query.Add("receiptsTopic", kafkaConfig.CommandReceiptsTopic)
	query.Add("changeEventsTopic", kafkaConfig.CommandReceiptsTopic)
	url := fmt.Sprintf("https://localhost:7159/api/File/GetKafkaConfiguration?%s", query.Encode())

	fmt.Println(url)

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	part, _ := writer.CreateFormFile("files", "accessKey.pem")
	io.Copy(part, kafkaConfig.AccessKey)

	part, _ = writer.CreateFormFile("files", "certificate.pem")
	io.Copy(part, kafkaConfig.Certificate)

	part, _ = writer.CreateFormFile("files", "ca.pem")
	io.Copy(part, kafkaConfig.Ca)

	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	req, err := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := httpClient.Do(req)

	body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(response.Header["Content-Type"])
	//fmt.Println("Ouput:")
	//fmt.Println(string(body))
	//ioutil.WriteFile("/Users/gh/Desktop/foo.zip", body, fs.FileMode(os.O_CREATE))
	return body
}
