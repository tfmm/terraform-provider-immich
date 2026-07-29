package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &albumResource{}
var _ resource.ResourceWithImportState = &albumResource{}

func NewAlbumResource() resource.Resource {
	return &albumResource{}
}

// albumResource defines the resource implementation.
type albumResource struct {
	client *client.Client
}

// albumResourceModel describes the resource data model.
type albumResourceModel struct {
	ID                    types.String     `tfsdk:"id"`
	Name                  types.String     `tfsdk:"name"`
	Description           types.String     `tfsdk:"description"`
	AlbumThumbnailAssetId types.String     `tfsdk:"album_thumbnail_asset_id"`
	IsActivityEnabled     types.Bool       `tfsdk:"is_activity_enabled"`
	Order                 types.String     `tfsdk:"order"`
	AssetIds              []types.String   `tfsdk:"asset_ids"`
	Users                 []albumUserModel `tfsdk:"users"`
}

type albumUserModel struct {
	UserId types.String `tfsdk:"user_id"`
	Role   types.String `tfsdk:"role"`
}

func (r *albumResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_album"
}

func (r *albumResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich album.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the album.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the album.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional description of the album.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"album_thumbnail_asset_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the asset used as the album's thumbnail.",
			},
			"is_activity_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether user activity (comments/likes) is enabled for this album.",
			},
			"order": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Sort order for assets in the album. Must be either `asc` or `desc`.",
			},
			"asset_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of asset IDs to include in the album.",
			},
			"users": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "List of users to share the album with.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Unique identifier of the user to share with.",
						},
						"role": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Role granted to the user. Must be either `editor` or `viewer`.",
						},
					},
				},
			},
		},
	}
}

func (r *albumResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func updateAlbumResourceModel(data *albumResourceModel, album *client.Album) {
	data.ID = types.StringValue(album.ID)
	data.Name = types.StringValue(album.AlbumName)

	if album.Description != nil && *album.Description != "" {
		data.Description = types.StringValue(*album.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.AlbumThumbnailAssetId = types.StringPointerValue(album.AlbumThumbnailAssetId)
	data.IsActivityEnabled = types.BoolValue(album.IsActivityEnabled)

	if album.Order != "" {
		data.Order = types.StringValue(album.Order)
	} else {
		data.Order = types.StringNull()
	}

	var users []albumUserModel
	for _, u := range album.AlbumUsers {
		if u.User != nil && !strings.EqualFold(u.Role, "owner") {
			users = append(users, albumUserModel{
				UserId: types.StringValue(u.User.ID),
				Role:   types.StringValue(u.Role),
			})
		}
	}
	if len(users) == 0 && (data.Users == nil || len(data.Users) == 0) {
		data.Users = nil
	} else {
		data.Users = users
	}
}

func (r *albumResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data albumResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var desc *string
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		d := data.Description.ValueString()
		desc = &d
	}

	createReq := client.CreateAlbumRequest{
		AlbumName:   data.Name.ValueString(),
		Description: desc,
	}

	for _, u := range data.Users {
		createReq.AlbumUsers = append(createReq.AlbumUsers, client.AlbumUserCreate{
			UserId: u.UserId.ValueString(),
			Role:   u.Role.ValueString(),
		})
	}

	for _, id := range data.AssetIds {
		createReq.AssetIds = append(createReq.AssetIds, id.ValueString())
	}

	album, err := r.client.CreateAlbum(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create album, got error: %s", err))
		return
	}

	updateAlbumResourceModel(&data, album)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *albumResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data albumResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	album, err := r.client.GetAlbum(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read album, got error: %s", err))
		return
	}

	updateAlbumResourceModel(&data, album)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *albumResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state albumResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var desc *string
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		d := plan.Description.ValueString()
		desc = &d
	} else if plan.Description.IsNull() {
		empty := ""
		desc = &empty
	}

	updateReq := client.UpdateAlbumRequest{
		AlbumName:             plan.Name.ValueString(),
		Description:           desc,
		AlbumThumbnailAssetId: plan.AlbumThumbnailAssetId.ValueStringPointer(),
		Order:                 plan.Order.ValueString(),
	}

	if !plan.IsActivityEnabled.IsNull() {
		enabled := plan.IsActivityEnabled.ValueBool()
		updateReq.IsActivityEnabled = &enabled
	}

	album, err := r.client.UpdateAlbum(plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update album info, got error: %s", err))
		return
	}

	updateAlbumResourceModel(&plan, album)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *albumResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data albumResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlbum(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete album, got error: %s", err))
		return
	}
}

func (r *albumResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
