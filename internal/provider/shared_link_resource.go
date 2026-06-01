package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/immich-app/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &sharedLinkResource{}
var _ resource.ResourceWithImportState = &sharedLinkResource{}

func NewSharedLinkResource() resource.Resource {
	return &sharedLinkResource{}
}

// sharedLinkResource defines the resource implementation.
type sharedLinkResource struct {
	client *client.Client
}

// sharedLinkResourceModel describes the resource data model.
type sharedLinkResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Type          types.String   `tfsdk:"type"`
	AssetIds      []types.String `tfsdk:"asset_ids"`
	AlbumId       types.String   `tfsdk:"album_id"`
	Description   types.String   `tfsdk:"description"`
	Password      types.String   `tfsdk:"password"`
	Slug          types.String   `tfsdk:"slug"`
	ExpiresAt     types.String   `tfsdk:"expires_at"`
	AllowUpload   types.Bool     `tfsdk:"allow_upload"`
	AllowDownload types.Bool     `tfsdk:"allow_download"`
	ShowMetadata  types.Bool     `tfsdk:"show_metadata"`
	Key           types.String   `tfsdk:"key"`
}

func (r *sharedLinkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_link"
}

func (r *sharedLinkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich shared link for albums or individual assets.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the shared link.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of the shared link. Must be either `ALBUM` or `INDIVIDUAL`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of asset IDs to share (required if type is `INDIVIDUAL`).",
			},
			"album_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the album to share (required if type is `ALBUM`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description for the shared link.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional password protection for the link.",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom URL slug for the shared link.",
			},
			"expires_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ISO 8601 formatted timestamp when the link expires.",
			},
			"allow_upload": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to allow users with the link to upload assets.",
			},
			"allow_download": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to allow users with the link to download assets.",
			},
			"show_metadata": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to show asset metadata to users with the link.",
			},
			"key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The encryption key for the shared link.",
			},
		},
	}
}

func (r *sharedLinkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sharedLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.SharedLinkCreateRequest{
		Type:          data.Type.ValueString(),
		Description:   data.Description.ValueStringPointer(),
		Password:      data.Password.ValueStringPointer(),
		Slug:          data.Slug.ValueStringPointer(),
		ExpiresAt:     data.ExpiresAt.ValueStringPointer(),
		AllowUpload:   data.AllowUpload.ValueBoolPointer(),
		AllowDownload: data.AllowDownload.ValueBoolPointer(),
		ShowMetadata:  data.ShowMetadata.ValueBoolPointer(),
	}

	if !data.AlbumId.IsNull() {
		albumId := data.AlbumId.ValueString()
		createReq.AlbumId = &albumId
	}

	if len(data.AssetIds) > 0 {
		assetIds := make([]string, len(data.AssetIds))
		for i, id := range data.AssetIds {
			assetIds[i] = id.ValueString()
		}
		createReq.AssetIds = assetIds
	}

	sharedLink, err := r.client.CreateSharedLink(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create shared link, got error: %s", err))
		return
	}

	data.ID = types.StringValue(sharedLink.ID)
	data.Key = types.StringValue(sharedLink.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sharedLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sharedLink, err := r.client.GetSharedLink(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read shared link, got error: %s", err))
		return
	}

	data.Description = types.StringPointerValue(sharedLink.Description)
	data.Type = types.StringValue(sharedLink.Type)
	data.ExpiresAt = types.StringPointerValue(sharedLink.ExpiresAt)
	data.AllowUpload = types.BoolValue(sharedLink.AllowUpload)
	data.AllowDownload = types.BoolValue(sharedLink.AllowDownload)
	data.ShowMetadata = types.BoolValue(sharedLink.ShowMetadata)
	data.Slug = types.StringPointerValue(sharedLink.Slug)
	data.Key = types.StringValue(sharedLink.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sharedLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.SharedLinkUpdateRequest{
		Description:   data.Description.ValueStringPointer(),
		Password:      data.Password.ValueStringPointer(),
		Slug:          data.Slug.ValueStringPointer(),
		ExpiresAt:     data.ExpiresAt.ValueStringPointer(),
		AllowUpload:   data.AllowUpload.ValueBoolPointer(),
		AllowDownload: data.AllowDownload.ValueBoolPointer(),
		ShowMetadata:  data.ShowMetadata.ValueBoolPointer(),
	}

	_, err := r.client.UpdateSharedLink(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update shared link, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sharedLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSharedLink(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete shared link, got error: %s", err))
		return
	}
}

func (r *sharedLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
