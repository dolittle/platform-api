package azure

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/Azure/azure-storage-file-go/azfile"
)

func CreateLink(accountName string, accountKey string, shareName string, filePath string, expireIn time.Time) (string, error) {

	// Use your Storage account's name and key to create a credential object; this is required to sign a SAS.
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", err
	}

	// Set the desired SAS signature values and sign them with the shared key credentials to get the SAS query parameters.
	sasQueryParams, err := azfile.FileSASSignatureValues{
		Protocol:   azfile.SASProtocolHTTPS, // Users MUST use HTTPS (not HTTP)
		ExpiryTime: expireIn,                // 48-hours before expiration
		ShareName:  shareName,
		FilePath:   filePath,

		// To produce a share SAS (as opposed to a file SAS in this example), assign to Permissions using
		// ShareSASPermissions and make sure the FilePath field is "" (the default).
		Permissions: azfile.FileSASPermissions{
			Read:  true,
			Write: false}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		return "", err
	}

	qp := sasQueryParams.Encode()
	urlToSendToSomeone := fmt.Sprintf("https://%s.file.core.windows.net/%s/%s?%s",
		accountName, shareName, filePath, qp)
	return urlToSendToSomeone, nil
}

func LatestX(accountName string, accountKey string, shareName string) (ListResponse, error) {
	// Use your Storage account's name and key to create a credential object; this is used to access your account.
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return ListResponse{}, err
	}

	p := azfile.NewPipeline(credential, azfile.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net", accountName))
	serviceURL := azfile.NewServiceURL(*u, p)

	ctx := context.Background()
	shareURL := serviceURL.NewShareURL(shareName)
	directoryURL := shareURL.NewDirectoryURL("mongo")

	found := make([]string, 0)
	for marker := (azfile.Marker{}); marker.NotDone(); {
		// Prefix is to strict
		listResponse, err := directoryURL.ListFilesAndDirectoriesSegment(ctx, marker, azfile.ListFilesAndDirectoriesOptions{})

		if err != nil {
			return ListResponse{}, err
		}

		marker = listResponse.NextMarker

		for _, fileEntry := range listResponse.FileItems {

			//found = append(found, fmt.Sprintf("%s/%s", directoryURL.String(), fileEntry.Name))
			found = append(found, fmt.Sprintf("%s/%s", directoryURL.URL().Path, fileEntry.Name))
			//found = append(found, fileEntry.Name)
		}
	}

	// Get the latest files, due to timestamp
	last := 20
	size := len(found)
	if size < last {
		last = size
	}
	latest := found[size-last:]
	// Flip the order and put the latest first
	sort.Sort(sort.Reverse(sort.StringSlice(latest)))

	return ListResponse{
		AccountName: accountName,
		ShareName:   shareName,
		Prefix:      serviceURL.String(),
		Files:       latest,
	}, nil
}

// EnsureFileShareExists tries to create a fileshare with a default quota in the given storage account.
// If the fileshare already exists it returns nil.
func EnsureFileShareExists(accountName, accountKey, shareName string) error {
	// Use your Storage account's name and key to create a credential object; this is used to access your account.
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)

	if err != nil {
		return err
	}

	pipeline := azfile.NewPipeline(credential, azfile.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net/%s", accountName, shareName))
	if err != nil {
		return err
	}

	shareURL := azfile.NewShareURL(*u, pipeline)
	ctx := context.Background()

	// NOTE: Metadata key names are always converted to lowercase before being sent to the Storage Service.
	// Therefore, you should always use lowercase letters; especially when querying a map for a metadata key.
	if _, err := shareURL.Create(ctx, azfile.Metadata{}, 0); err != nil {
		if storageErr, ok := err.(azfile.StorageError); ok && storageErr.ServiceCode() == "ShareAlreadyExists" {
			return nil
		}
		return err
	}

	return nil
}
