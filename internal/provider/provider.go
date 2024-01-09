package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*netboxProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &netboxProvider{}
	}
}

type netboxProvider struct{}

func (p *netboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {

}

func (p *netboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

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
