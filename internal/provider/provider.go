package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/immich-app/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &immichProvider{}

// immichProvider is the provider implementation.
type immichProvider struct {
	version string
}

// immichProviderModel describes the provider data model.
type immichProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *immichProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "immich"
	resp.Version = p.version
}

func (p *immichProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Immich provider is used to manage resources on an Immich server.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The full URL of the Immich API endpoint. Can also be set via the `IMMICH_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authenticating with the Immich server. Can also be set via the `IMMICH_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *immichProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data immichProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("IMMICH_ENDPOINT")
	apiKey := os.Getenv("IMMICH_API_KEY")

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError("Missing Immich API Endpoint", "The provider cannot create the Immich API client without an endpoint.")
		return
	}

	if apiKey == "" {
		resp.Diagnostics.AddError("Missing Immich API Key", "The provider cannot create the Immich API client without an API key.")
		return
	}

	c := client.NewClient(endpoint, apiKey)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *immichProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewApiKeyResource,
		NewSharedLinkResource,
		NewAlbumResource,
		NewSystemConfigResource,
	}
}

func (p *immichProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUsersDataSource,
		NewAlbumsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &immichProvider{
			version: version,
		}
	}
}
