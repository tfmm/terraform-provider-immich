package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &librariesDataSource{}

func NewLibrariesDataSource() datasource.DataSource {
	return &librariesDataSource{}
}

// librariesDataSource defines the data source implementation.
type librariesDataSource struct {
	client *client.Client
}

// librariesDataSourceModel describes the data source data model.
type librariesDataSourceModel struct {
	Libraries []librariesModel `tfsdk:"libraries"`
}

type librariesModel struct {
	ID         types.String `tfsdk:"id"`
	OwnerId    types.String `tfsdk:"owner_id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	AssetCount types.Int64  `tfsdk:"asset_count"`
}

func (d *librariesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_libraries"
}

func (d *librariesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of all Immich libraries.",

		Attributes: map[string]schema.Attribute{
			"libraries": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of libraries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the library.",
						},
						"owner_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier of the library owner.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Display name of the library.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of the library (UPLOAD or EXTERNAL).",
						},
						"asset_count": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of assets in the library.",
						},
					},
				},
			},
		},
	}
}

func (d *librariesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *librariesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data librariesDataSourceModel

	libraries, err := d.client.GetLibraries()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read libraries, got error: %s", err))
		return
	}

	for _, library := range libraries {
		libraryState := librariesModel{
			ID:         types.StringValue(library.ID),
			OwnerId:    types.StringValue(library.OwnerId),
			Name:       types.StringValue(library.Name),
			Type:       types.StringValue(library.Type),
			AssetCount: types.Int64Value(int64(library.AssetCount)),
		}
		data.Libraries = append(data.Libraries, libraryState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
