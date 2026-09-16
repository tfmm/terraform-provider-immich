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
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &sharedLinkResource{}
var _ resource.ResourceWithImportState = &sharedLinkResource{}
var _ resource.ResourceWithValidateConfig = &sharedLinkResource{}

func NewSharedLinkResource() resource.Resource {
	return &sharedLinkResource{}
}

// sharedLinkResource defines the resource implementation.
type sharedLinkResource struct {
	client *client.Client
}

// sharedLinkResourceModel describes the resource data model.
type sharedLinkResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Type              types.String   `tfsdk:"type"`
	AssetIds          []types.String `tfsdk:"asset_ids"`
	AlbumId           types.String   `tfsdk:"album_id"`
	Description       types.String   `tfsdk:"description"`
	Password          types.String   `tfsdk:"password"`
	PasswordWO        types.String   `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64    `tfsdk:"password_wo_version"`
	Slug              types.String   `tfsdk:"slug"`
	ExpiresAt         types.String   `tfsdk:"expires_at"`
	AllowUpload       types.Bool     `tfsdk:"allow_upload"`
	AllowDownload     types.Bool     `tfsdk:"allow_download"`
	ShowMetadata      types.Bool     `tfsdk:"show_metadata"`
	Key               types.String   `tfsdk:"key"`
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
				DeprecationMessage:  "Use `password_wo` instead, which is never persisted to plan or state.",
				MarkdownDescription: "Optional password protection for the link. Persisted to state in plain text (aside from standard state encryption); mutually exclusive with `password_wo`. Deprecated in favor of `password_wo`.",
			},
			"password_wo": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				MarkdownDescription: "Write-only password protection for the link. Never persisted to plan or state. Requires Terraform 1.11+. Must be paired with `password_wo_version`; bump the version to rotate the password on a later apply. Mutually exclusive with `password`.",
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Arbitrary version number for `password_wo`. Increment it to signal that the password should change; the number itself has no meaning beyond change detection, since `password_wo`'s value is never stored to compare against.",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *sharedLinkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data sharedLinkResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasPassword := !data.Password.IsNull() && !data.Password.IsUnknown()
	hasPasswordWO := !data.PasswordWO.IsNull() && !data.PasswordWO.IsUnknown()

	if hasPassword && hasPasswordWO {
		resp.Diagnostics.AddError("Conflicting Password Attributes", "Only one of `password` or `password_wo` may be set.")
	}
	if hasPasswordWO && data.PasswordWOVersion.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("password_wo_version"), "Missing Password Version", "`password_wo_version` must be set when `password_wo` is used, and bumped to rotate the password.")
	}
}

func (r *sharedLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config sharedLinkResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := data.Password.ValueStringPointer()
	if !config.PasswordWO.IsNull() {
		p := config.PasswordWO.ValueString()
		password = &p
	}

	createReq := client.SharedLinkCreateRequest{
		Type:          data.Type.ValueString(),
		Description:   data.Description.ValueStringPointer(),
		Password:      password,
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

	sharedLink, err := r.client.CreateSharedLink(ctx, createReq)
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

	sharedLink, err := r.client.GetSharedLink(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
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
	var plan, state, config sharedLinkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.SharedLinkUpdateRequest{
		Description:   plan.Description.ValueStringPointer(),
		Slug:          plan.Slug.ValueStringPointer(),
		ExpiresAt:     plan.ExpiresAt.ValueStringPointer(),
		AllowUpload:   plan.AllowUpload.ValueBoolPointer(),
		AllowDownload: plan.AllowDownload.ValueBoolPointer(),
		ShowMetadata:  plan.ShowMetadata.ValueBoolPointer(),
	}

	if !config.PasswordWO.IsNull() {
		// The write-only value itself is never available to compare against
		// a prior apply, so a bump in password_wo_version is the only
		// signal that the password should be rotated.
		if !plan.PasswordWOVersion.Equal(state.PasswordWOVersion) {
			p := config.PasswordWO.ValueString()
			updateReq.Password = &p
		}
	} else {
		updateReq.Password = plan.Password.ValueStringPointer()
	}

	_, err := r.client.UpdateSharedLink(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update shared link, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sharedLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sharedLinkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSharedLink(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete shared link, got error: %s", err))
		return
	}
}

func (r *sharedLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
