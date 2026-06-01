package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &assetResource{}
var _ resource.ResourceWithImportState = &assetResource{}

func NewAssetResource() resource.Resource {
	return &assetResource{}
}

// assetResource defines the resource implementation.
type assetResource struct {
	client *client.Client
}

// assetResourceModel describes the resource data model.
type assetResourceModel struct {
	ID          types.String  `tfsdk:"id"`
	IsFavorite  types.Bool    `tfsdk:"is_favorite"`
	IsArchived  types.Bool    `tfsdk:"is_archived"`
	Description types.String  `tfsdk:"description"`
	Latitude    types.Float64 `tfsdk:"latitude"`
	Longitude   types.Float64 `tfsdk:"longitude"`
}

func (r *assetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (r *assetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich asset's metadata. Note: This resource does not support uploading files. It is intended for managing metadata of existing assets.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier for the asset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_favorite": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the asset is marked as a favorite.",
			},
			"is_archived": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the asset is archived.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the asset.",
			},
			"latitude": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Latitude of the asset.",
			},
			"longitude": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Longitude of the asset.",
			},
		},
	}
}

func (r *assetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *assetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Since we don't support POST /assets (upload), Create will just update the metadata of an existing asset ID
	// provided in the config.
	updateReq := client.UpdateAssetRequest{}
	if !data.IsFavorite.IsNull() {
		val := data.IsFavorite.ValueBool()
		updateReq.IsFavorite = &val
	}
	if !data.IsArchived.IsNull() {
		val := data.IsArchived.ValueBool()
		updateReq.IsArchived = &val
	}
	if !data.Description.IsNull() {
		updateReq.Description = data.Description.ValueString()
	}
	if !data.Latitude.IsNull() {
		val := data.Latitude.ValueFloat64()
		updateReq.Latitude = &val
	}
	if !data.Longitude.IsNull() {
		val := data.Longitude.ValueFloat64()
		updateReq.Longitude = &val
	}

	_, err := r.client.UpdateAsset(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update asset metadata, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	asset, err := r.client.GetAsset(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read asset, got error: %s", err))
		return
	}

	data.IsFavorite = types.BoolValue(asset.IsFavorite)
	data.IsArchived = types.BoolValue(asset.IsArchived)
	data.Description = types.StringValue(asset.Description)
	if asset.ExifInfo != nil {
		data.Latitude = types.Float64Value(asset.ExifInfo.Latitude)
		data.Longitude = types.Float64Value(asset.ExifInfo.Longitude)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateAssetRequest{}
	if !data.IsFavorite.IsNull() {
		val := data.IsFavorite.ValueBool()
		updateReq.IsFavorite = &val
	}
	if !data.IsArchived.IsNull() {
		val := data.IsArchived.ValueBool()
		updateReq.IsArchived = &val
	}
	if !data.Description.IsNull() {
		updateReq.Description = data.Description.ValueString()
	}
	if !data.Latitude.IsNull() {
		val := data.Latitude.ValueFloat64()
		updateReq.Latitude = &val
	}
	if !data.Longitude.IsNull() {
		val := data.Longitude.ValueFloat64()
		updateReq.Longitude = &val
	}

	_, err := r.client.UpdateAsset(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update asset, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAssets([]string{data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete asset, got error: %s", err))
		return
	}
}

func (r *assetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
