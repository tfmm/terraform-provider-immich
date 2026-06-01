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
var _ datasource.DataSource = &albumsDataSource{}

func NewAlbumsDataSource() datasource.DataSource {
	return &albumsDataSource{}
}

// albumsDataSource defines the data source implementation.
type albumsDataSource struct {
	client *client.Client
}

// albumsDataSourceModel describes the data source data model.
type albumsDataSourceModel struct {
	Albums []albumsModel `tfsdk:"albums"`
}

type albumsModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AssetCount  types.Int64  `tfsdk:"asset_count"`
}

func (d *albumsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_albums"
}

func (d *albumsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of all Immich albums.",

		Attributes: map[string]schema.Attribute{
			"albums": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of albums.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the album.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Display name of the album.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of the album.",
						},
						"asset_count": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of assets in the album.",
						},
					},
				},
			},
		},
	}
}

func (d *albumsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *albumsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data albumsDataSourceModel

	albums, err := d.client.GetAlbums()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read albums, got error: %s", err))
		return
	}

	for _, album := range albums {
		albumState := albumsModel{
			ID:          types.StringValue(album.ID),
			Name:        types.StringValue(album.AlbumName),
			Description: types.StringValue(album.Description),
			AssetCount:  types.Int64Value(int64(album.AssetCount)),
		}
		data.Albums = append(data.Albums, albumState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
