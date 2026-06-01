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
var _ resource.Resource = &activityResource{}
var _ resource.ResourceWithImportState = &activityResource{}

func NewActivityResource() resource.Resource {
	return &activityResource{}
}

// activityResource defines the resource implementation.
type activityResource struct {
	client *client.Client
}

// activityResourceModel describes the resource data model.
type activityResourceModel struct {
	ID        types.String `tfsdk:"id"`
	AlbumId   types.String `tfsdk:"album_id"`
	AssetId   types.String `tfsdk:"asset_id"`
	Type      types.String `tfsdk:"type"`
	Comment   types.String `tfsdk:"comment"`
	CreatedAt types.String `tfsdk:"created_at"`
	UserId    types.String `tfsdk:"user_id"`
}

func (r *activityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity"
}

func (r *activityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich activity (comment or like).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the activity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"album_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the album.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"asset_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the asset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of activity (comment or like).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Comment text (required for type 'comment').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the activity was created.",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the user who performed the activity.",
			},
		},
	}
}

func (r *activityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *activityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data activityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateActivityRequest{
		Type:    data.Type.ValueString(),
		AlbumId: data.AlbumId.ValueString(),
		AssetId: data.AssetId.ValueString(),
		Comment: data.Comment.ValueString(),
	}

	activity, err := r.client.CreateActivity(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create activity, got error: %s", err))
		return
	}

	data.ID = types.StringValue(activity.ID)
	data.CreatedAt = types.StringValue(activity.CreatedAt)
	data.UserId = types.StringValue(activity.User.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *activityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data activityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// We have to find the activity in the list
	activities, err := r.client.GetActivities(data.AlbumId.ValueString(), data.AssetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read activities, got error: %s", err))
		return
	}

	var found *client.Activity
	for _, a := range activities {
		if a.ID == data.ID.ValueString() {
			found = &a
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Type = types.StringValue(strings.ToLower(found.Type))
	data.Comment = types.StringValue(found.Comment)
	data.CreatedAt = types.StringValue(found.CreatedAt)
	data.UserId = types.StringValue(found.User.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *activityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Immich doesn't seem to support updating activities.
}

func (r *activityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data activityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteActivity(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete activity, got error: %s", err))
		return
	}
}

func (r *activityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: album_id/activity_id or album_id/asset_id/activity_id
	idParts := strings.Split(req.ID, "/")
	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("album_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
	} else if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("album_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("asset_id"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[2])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: album_id/activity_id or album_id/asset_id/activity_id. Got: %q", req.ID),
		)
	}
}
