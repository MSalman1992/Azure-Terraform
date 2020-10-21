package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/servicefabricmesh/mgmt/2018-09-01-preview/servicefabricmesh"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	ApplicationClient *servicefabricmesh.ApplicationClient
	GatewayClient     *servicefabricmesh.GatewayClient
	NetworkClient     *servicefabricmesh.NetworkClient
	SecretClient      *servicefabricmesh.SecretClient
	SecretValueClient *servicefabricmesh.SecretValueClient
	ServiceClient     *servicefabricmesh.ServiceClient
}

func NewClient(o *common.ClientOptions) *Client {
	applicationsClient := servicefabricmesh.NewApplicationClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&applicationsClient.Client, o.ResourceManagerAuthorizer)

	gatewaysClient := servicefabricmesh.NewGatewayClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&gatewaysClient.Client, o.ResourceManagerAuthorizer)

	networksClient := servicefabricmesh.NewNetworkClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&networksClient.Client, o.ResourceManagerAuthorizer)

	secretsClient := servicefabricmesh.NewSecretClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&secretsClient.Client, o.ResourceManagerAuthorizer)

	secretValuesClient := servicefabricmesh.NewSecretValueClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&secretValuesClient.Client, o.ResourceManagerAuthorizer)

	servicesClient := servicefabricmesh.NewServiceClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&servicesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		ApplicationClient: &applicationsClient,
		GatewayClient:     &gatewaysClient,
		NetworkClient:     &networksClient,
		SecretClient:      &secretsClient,
		SecretValueClient: &secretValuesClient,
		ServiceClient:     &servicesClient,
	}
}
