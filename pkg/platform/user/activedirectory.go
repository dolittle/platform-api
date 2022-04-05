package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

type UserActiveDirectory interface {
	AddUserToGroupByEmail(email string, groupID string) error
	AddUserToGroup(userID string, groupID string) error
	RemoveUserFromGroupByEmail(email string, groupID string) error
	RemoveUserFromGroup(userID string, groupID string) error
	GetUserIDByEmail(email string) (string, error)
	GetUsersInApplication(groupID string) ([]ActiveDirectoryUserInfo, error)
	GetGroupIDByApplicationID(applicationID string) (string, error)
}

type ActiveDirectoryUserInfo struct {
	ID    string
	Email string
}

type activeDirectoryClient struct {
	groupClient graphrbac.GroupsClient
	userClient  graphrbac.UsersClient
}

func NewUserActiveDirectory(groupClient graphrbac.GroupsClient, userClient graphrbac.UsersClient) activeDirectoryClient {
	return activeDirectoryClient{
		groupClient: groupClient,
		userClient:  userClient,
	}
}

func (c activeDirectoryClient) GetUsersInApplication(groupID string) ([]ActiveDirectoryUserInfo, error) {
	users := make([]ActiveDirectoryUserInfo, 0)
	ctx := context.Background()
	pager, err := c.groupClient.GetGroupMembers(ctx, groupID)

	if err != nil {
		return users, err
	}

	for _, v := range pager.Values() {
		user, success := v.AsUser()
		if !success {
			continue
		}

		mail := ""
		if user.Mail != nil {
			mail = *user.Mail
		}

		users = append(users, ActiveDirectoryUserInfo{
			ID:    *user.ObjectID,
			Email: mail,
		})
	}

	return users, nil
}

func (c activeDirectoryClient) GetGroupIDByApplicationID(applicationID string) (string, error) {
	ctx := context.TODO()
	filter := fmt.Sprintf("displayName eq 'Application-%s'", applicationID)
	r, err := c.groupClient.List(ctx, filter)

	if err != nil {
		fmt.Println("err looking up", err)
		// issue.connecting
		return "", err
	}

	items := r.Values()
	size := len(items)

	if size == 0 {
		return "", ErrNotFound
	}

	if size != 1 {
		return "", ErrTooManyResults
	}

	aGroup, success := items[0].AsADGroup()
	if !success {
		return "", ErrNotFound
	}

	return *aGroup.ObjectID, nil
}

func (c activeDirectoryClient) GetUserIDByEmail(email string) (string, error) {
	ctx := context.TODO()
	filter := fmt.Sprintf("mail eq '%s'", email)
	r, err := c.userClient.List(ctx, filter, "")

	if err != nil {
		// issue.connecting
		return "", err
	}

	items := r.Values()
	size := len(items)

	if size == 0 {
		return "", ErrNotFound
	}

	if size != 1 {
		return "", ErrTooManyResults
	}

	user, success := items[0].AsUser()
	if !success {
		return "", ErrNotFound
	}

	return *user.ObjectID, nil
}

func (c activeDirectoryClient) AddUserToGroupByEmail(email string, groupID string) error {
	userID, err := c.GetUserIDByEmail(email)
	if err != nil {
		return err
	}
	return c.AddUserToGroup(userID, groupID)
}

// TODO I might need to talk email, as today its the only connection between kratos + AAD
// AddUserToGroup
func (c activeDirectoryClient) AddUserToGroup(userID string, groupID string) error {
	groupObjectID := groupID

	client := c.groupClient

	tenantID := client.TenantID
	memberObjectID := userID
	ctx := context.Background()
	result, err := client.AddMember(ctx, groupObjectID, graphrbac.GroupAddMemberParameters{
		URL: to.StringPtr(fmt.Sprintf(
			"https://graph.microsoft.com/%s/users/%s",
			tenantID,
			memberObjectID,
		)),
	})

	// 204 == success
	// 400 = already exists
	// One or more added object references already exist for the following modified properties: 'members'. message.value
	// {"odata.error":{"code":"Request_BadRequest","message":{"lang":"en","value":"One or more added object references already exist for the following modified properties: 'members'."},"requestId":"3fd92770-5657-4033-8da3-4585aae4a338","date":"2022-02-10T22:13:27"}}
	if result.Response.StatusCode == http.StatusNoContent {
		return nil
	}

	fmt.Println(result.Response.StatusCode, err)
	defer result.Response.Body.Close()

	bodyBytes, err := io.ReadAll(result.Response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	if result.Response.StatusCode == http.StatusBadRequest {
		if strings.Contains(bodyString, "One or more added object references already exist for the following modified properties") {
			return ErrEmailAlreadyExists
		}
	}

	fmt.Println(bodyString)
	// TODO these errors need more work
	return errors.New("failed to add")
}

func (c activeDirectoryClient) RemoveUserFromGroupByEmail(email string, groupID string) error {
	userID, err := c.GetUserIDByEmail(email)
	if err != nil {
		return err
	}
	return c.RemoveUserFromGroup(userID, groupID)
}

// RemoveUserFromGroup
func (c activeDirectoryClient) RemoveUserFromGroup(userID string, groupID string) error {
	groupObjectID := groupID
	memberObjectID := userID
	client := c.groupClient
	ctx := context.Background()
	result, err := client.RemoveMember(ctx, groupObjectID, memberObjectID)
	// 204 == success
	// 404 == not found
	if result.Response.StatusCode == http.StatusNoContent {
		return nil
	}
	fmt.Println(result.Response.StatusCode, err)
	defer result.Response.Body.Close()

	bodyBytes, err := io.ReadAll(result.Response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
	// TODO these errors need more work
	return errors.New("failed to remove")
}

func NewBearerAuthorizerFromEnvironmentVariables(settings auth.EnvironmentSettings) *autorest.BearerAuthorizer {
	cred, err := settings.GetClientCredentials()
	if err != nil {
		panic(err)
	}

	cloudName := "AzurePublicCloud"
	env, _ := azure.EnvironmentFromName(cloudName)
	oauthConfig, err := adal.NewOAuthConfig(
		env.ActiveDirectoryEndpoint, cred.TenantID)
	if err != nil {
		panic(err)
	}

	token, err := adal.NewServicePrincipalToken(
		*oauthConfig, cred.ClientID, cred.ClientSecret, env.GraphEndpoint)
	if err != nil {
		panic(err)
	}
	return autorest.NewBearerAuthorizer(token)
}

func NewUserClient(tenantID string, authorizer *autorest.BearerAuthorizer) graphrbac.UsersClient {
	userClient := graphrbac.NewUsersClient(tenantID)
	userClient.Authorizer = authorizer
	userClient.AddToUserAgent("dolittle")
	return userClient
}

func NewGroupsClient(tenantID string, authorizer *autorest.BearerAuthorizer) graphrbac.GroupsClient {
	groupsClient := graphrbac.NewGroupsClient(tenantID)
	groupsClient.Authorizer = authorizer
	groupsClient.AddToUserAgent("dolittle")
	return groupsClient
}
