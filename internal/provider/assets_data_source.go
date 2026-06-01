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
var _ datasource.DataSource = &assetsDataSource{}

func NewAssetsDataSource() datasource.DataSource {
	return &assetsDataSource{}
}

// assetsDataSource defines the data source implementation.
type assetsDataSource struct {
	client *client.Client
}

// assetsDataSourceModel describes the data source data model.
type assetsDataSourceModel struct {
	IsFavorite       types.Bool    `tfsdk:"is_favorite"`
	Type             types.String  `tfsdk:"type"`
	OriginalFileName types.String  `tfsdk:"original_file_name"`
	City             types.String  `tfsdk:"city"`
	Country          types.String  `tfsdk:"country"`
	Make             types.String  `tfsdk:"make"`
	Model            types.String  `tfsdk:"model"`
	Assets           []assetsModel `tfsdk:"assets"`
}

type assetsModel struct {
	ID               types.String `tfsdk:"id"`
	OriginalFileName types.String `tfsdk:"original_file_name"`
	Type             types.String `tfsdk:"type"`
	IsFavorite       types.Bool   `tfsdk:"is_favorite"`
	IsArchived       types.Bool   `tfsdk:"is_archived"`
	FileCreatedAt    types.String `tfsdk:"file_created_at"`
}

func (d *assetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets"
}

func (d *assetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of assets based on search criteria.",

		Attributes: map[string]schema.Attribute{
			"is_favorite": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Filter for favorite assets.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by asset type (IMAGE or VIDEO).",
			},
			"original_file_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by original filename.",
			},
			"city": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by city.",
			},
			"country": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by country.",
			},
			"make": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by camera make.",
			},
			"model": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by camera model.",
			},
			"assets": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of assets matching the criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the asset.",
						},
						"original_file_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Original filename.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Asset type.",
						},
						"is_favorite": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the asset is a favorite.",
						},
						"is_archived": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the asset is archived.",
						},
						"file_created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "When the file was created.",
						},
					},
				},
			},
		},
	}
}

func (d *assetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *assetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data assetsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	searchReq := client.SearchAssetsRequest{
		Type:             data.Type.ValueString(),
		OriginalFileName: data.OriginalFileName.ValueString(),
		City:             data.City.ValueString(),
		Country:          data.Country.ValueString(),
		Make:             data.Make.ValueString(),
		Model:            data.Model.ValueString(),
		WithExif:         true,
	}

	if !data.IsFavorite.IsNull() {
		val := data.IsFavorite.ValueBool()
		searchReq.IsFavorite = &val
	}

	response, err := d.client.SearchAssets(searchReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search assets, got error: %s", err))
		return
	}

	data.Assets = []assetsModel{}
	for _, a := range response.Assets.Items {
		assetState := assetsModel{
			ID:               types.StringValue(a.ID),
			OriginalFileName: types.StringValue(a.OriginalFileName),
			Type:             types.StringValue(a.Type),
			IsFavorite:       types.BoolValue(a.IsFavorite),
			IsArchived:       types.BoolValue(a.IsArchived),
			FileCreatedAt:    types.StringValue(a.FileCreatedAt),
		}
		data.Assets = append(data.Assets, assetState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
