package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/e-breuninger/terraform-provider-netbox/internal/generate/provider_netbox"
	netboxclient "github.com/fbreckle/go-netbox/netbox/client"
	_ "github.com/fbreckle/go-netbox/netbox/client/status"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/goware/urlx"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	log "github.com/sirupsen/logrus"
)

var _ provider.Provider = (*netboxProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &netboxProvider{}
	}
}

type netboxProvider struct{}

func (p *netboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider_netbox.NetboxProviderSchema(ctx)
}

func (p *netboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config provider_netbox.NetboxModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ServerUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Unknown NetBox Server URL",
			"The provider cannot create the NetBox API client as there is an unknown configuration value for the NetBox Server URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_SERVER_URL environment variable.",
		)
	}

	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown NetBox API Token",
			"The provider cannot create the NetBox API client as there is an unknown configuration value for the NetBox API Token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_API_TOKEN environment variable.",
		)
	}

	if config.StripTrailingSlashesFromUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("strip_trailing_slashes_from_url"),
			"wip", "wip",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	serverUrl := os.Getenv("NETBOX_SERVER_URL")
	apiToken := os.Getenv("NETBOX_API_TOKEN")

	if !config.ServerUrl.IsNull() {
		serverUrl = config.ServerUrl.ValueString()
	}

	if serverUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Missing NetBox Server URL",
			"The provider cannot create the NetBox API client as there is a missing configuration value for the NetBox Server URL. "+
				"Set the host value in the configuration or use the NETBOX_SERVER_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing NetBox API Token",
			"The provider cannot create the NetBox API client as there is a missing configuration value for the NetBox API Token. "+
				"Set the host value in the configuration or use the NETBOX_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	stripTrailingSlashesFromUrlEnv := os.Getenv("NETBOX_STRIP_TRAILING_SLASHES_FROM_URL")
	var stripTrailingSlashesFromUrl bool = true
	if stripTrailingSlashesFromUrlEnv == "false" {
		stripTrailingSlashesFromUrl = false
	}

	if !config.StripTrailingSlashesFromUrl.IsNull() {
		stripTrailingSlashesFromUrl = config.StripTrailingSlashesFromUrl.ValueBool()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// End boilerplate part

	// Unless explicitly switched off, strip trailing slashes from the server url
	// Trailing slashes cause errors as seen in https://github.com/e-breuninger/terraform-provider-netbox/issues/198
	// and https://github.com/e-breuninger/terraform-provider-netbox/issues/300

	if stripTrailingSlashesFromUrl {
		trimmed := false

		// This is Go's poor man's while loop
		for strings.HasSuffix(serverUrl, "/") {
			serverUrl = strings.TrimRight(serverUrl, "/")
			trimmed = true
		}
		if trimmed {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("strip_trailing_slashes_from_url"),
				"Stripped trailing slashes from the `server_url` parameter",
				"Trailing slashes in the `server_url` parameter lead to problems in most setups, so all trailing slashes were stripped. Use the `strip_trailing_slashes_from_url` parameter to disable this feature or remove all trailing slashes in the `server_url` to disable this warning.",
			)
		}
	}

	// Create a new NetBox client using the configuration values
	log.WithFields(log.Fields{
		"server_url": serverUrl,
	}).Debug("Initializing Netbox client")

	if apiToken == "" {
		fmt.Errorf("missing netbox API key")
	}

	// parse serverUrl
	parsedURL, urlParseError := urlx.Parse(serverUrl)
	if urlParseError != nil {
		fmt.Errorf("error while trying to parse URL: %s", urlParseError)
	}

	desiredRuntimeClientSchemes := []string{parsedURL.Scheme}
	log.WithFields(log.Fields{
		"host":    parsedURL.Host,
		"schemes": desiredRuntimeClientSchemes,
	}).Debug("Initializing Netbox Open API runtime client")

	// build http client
	clientOpts := httptransport.TLSClientOptions{
		InsecureSkipVerify: true, // wip
	}

	trans, err := httptransport.TLSTransport(clientOpts)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	//	if cfg.Headers != nil && len(cfg.Headers) > 0 {
	//		log.WithFields(log.Fields{
	//			"custom_headers": cfg.Headers,
	//		}).Debug("Setting custom headers on every request to Netbox")
	//
	//		trans = customHeaderTransport{
	//			original: trans,
	//			headers:  cfg.Headers,
	//		}
	//	}

	httpClient := &http.Client{
		Transport: trans,
		Timeout:   time.Second * time.Duration(10), // tmp
	}

	transport := httptransport.NewWithClient(parsedURL.Host, parsedURL.Path+netboxclient.DefaultBasePath, desiredRuntimeClientSchemes, httpClient)
	//transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", fmt.Sprintf("Token %v", cfg.APIToken)) // tmp
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", fmt.Sprintf("Token %v", apiToken))
	transport.SetLogger(log.StandardLogger())
	client := netboxclient.New(transport, nil)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *netboxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netbox"
}

func (p *netboxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *netboxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
