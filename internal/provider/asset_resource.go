package provider

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

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
	ID             types.String  `tfsdk:"id"`
	Filename       types.String  `tfsdk:"filename"`
	DeviceId       types.String  `tfsdk:"device_id"`
	DeviceAssetId  types.String  `tfsdk:"device_asset_id"`
	FileCreatedAt  types.String  `tfsdk:"file_created_at"`
	FileModifiedAt types.String  `tfsdk:"file_modified_at"`
	IsFavorite     types.Bool    `tfsdk:"is_favorite"`
	IsArchived     types.Bool    `tfsdk:"is_archived"`
	Description    types.String  `tfsdk:"description"`
	Latitude       types.Float64 `tfsdk:"latitude"`
	Longitude      types.Float64 `tfsdk:"longitude"`
}

func (r *assetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (r *assetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich asset. Supports uploading a file or managing metadata of an existing asset.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Unique identifier for the asset. If filename is provided, this is computed after upload.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"filename": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to the local file to upload. If provided, the asset will be uploaded to Immich.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Unique identifier for the device/client. Defaults to 'terraform'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_asset_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Unique identifier for the asset on the device. Defaults to the filename.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_created_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp of when the file was created. Defaults to the current time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_modified_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ISO 8601 timestamp of when the file was last modified. Defaults to the current time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

	if !data.Filename.IsNull() {
		filename := data.Filename.ValueString()
		deviceId := "terraform"
		if !data.DeviceId.IsNull() {
			deviceId = data.DeviceId.ValueString()
		}
		deviceAssetId := filepath.Base(filename)
		if !data.DeviceAssetId.IsNull() {
			deviceAssetId = data.DeviceAssetId.ValueString()
		}

		fileCreatedAt := time.Now()
		if !data.FileCreatedAt.IsNull() {
			t, err := time.Parse(time.RFC3339, data.FileCreatedAt.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Time Parse Error", fmt.Sprintf("Unable to parse file_created_at: %s", err))
				return
			}
			fileCreatedAt = t
		}

		fileModifiedAt := time.Now()
		if !data.FileModifiedAt.IsNull() {
			t, err := time.Parse(time.RFC3339, data.FileModifiedAt.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Time Parse Error", fmt.Sprintf("Unable to parse file_modified_at: %s", err))
				return
			}
			fileModifiedAt = t
		}

		isFavorite := false
		if !data.IsFavorite.IsNull() {
			isFavorite = data.IsFavorite.ValueBool()
		}

		asset, err := r.client.UploadAsset(filename, deviceId, deviceAssetId, fileCreatedAt, fileModifiedAt, isFavorite)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to upload asset, got error: %s", err))
			return
		}

		data.ID = types.StringValue(asset.ID)
		data.DeviceId = types.StringValue(asset.DeviceId)
		data.DeviceAssetId = types.StringValue(asset.DeviceAssetId)
		data.FileCreatedAt = types.StringValue(asset.FileCreatedAt)
		data.FileModifiedAt = types.StringValue(asset.FileModifiedAt)
	}

	if data.ID.IsNull() {
		resp.Diagnostics.AddError("Config Error", "Either 'id' or 'filename' must be provided.")
		return
	}

	// Update metadata
	updateReq := client.UpdateAssetRequest{}
	needsUpdate := false

	if !data.IsFavorite.IsNull() {
		val := data.IsFavorite.ValueBool()
		updateReq.IsFavorite = &val
		needsUpdate = true
	}
	if !data.IsArchived.IsNull() {
		val := data.IsArchived.ValueBool()
		updateReq.IsArchived = &val
		needsUpdate = true
	}
	if !data.Description.IsNull() {
		updateReq.Description = data.Description.ValueString()
		needsUpdate = true
	}
	if !data.Latitude.IsNull() {
		val := data.Latitude.ValueFloat64()
		updateReq.Latitude = &val
		needsUpdate = true
	}
	if !data.Longitude.IsNull() {
		val := data.Longitude.ValueFloat64()
		updateReq.Longitude = &val
		needsUpdate = true
	}

	if needsUpdate {
		_, err := r.client.UpdateAsset(data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update asset metadata, got error: %s", err))
			return
		}
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

	data.ID = types.StringValue(asset.ID)
	data.IsFavorite = types.BoolValue(asset.IsFavorite)
	data.IsArchived = types.BoolValue(asset.IsArchived)
	data.Description = types.StringValue(asset.Description)
	data.DeviceId = types.StringValue(asset.DeviceId)
	data.DeviceAssetId = types.StringValue(asset.DeviceAssetId)
	data.FileCreatedAt = types.StringValue(asset.FileCreatedAt)
	data.FileModifiedAt = types.StringValue(asset.FileModifiedAt)

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
