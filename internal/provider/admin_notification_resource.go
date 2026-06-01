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
var _ resource.Resource = &adminNotificationResource{}
var _ resource.ResourceWithImportState = &adminNotificationResource{}

func NewAdminNotificationResource() resource.Resource {
	return &adminNotificationResource{}
}

// adminNotificationResource defines the resource implementation.
type adminNotificationResource struct {
	client *client.Client
}

// adminNotificationResourceModel describes the resource data model.
type adminNotificationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Level       types.String `tfsdk:"level"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

func (r *adminNotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_notification"
}

func (r *adminNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sends a system-wide notification as an administrator. Note: This resource triggers a notification upon creation. Deleting the resource will attempt to remove the notification.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the created notification.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of notification (e.g. `SYSTEM`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"level": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Severity level (`INFO`, `WARNING`, `ERROR`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Notification title.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Notification message body.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *adminNotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adminNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data adminNotificationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateAdminNotificationRequest{
		Type:        data.Type.ValueString(),
		Level:       data.Level.ValueString(),
		Title:       data.Title.ValueString(),
		Description: data.Description.ValueString(),
	}

	notification, err := r.client.CreateAdminNotification(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send admin notification, got error: %s", err))
		return
	}

	data.ID = types.StringValue(notification.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *adminNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Notifications are usually transient and might not be easily refreshable by ID alone
	// if they disappear from the system.
	// We'll keep what's in state unless we can find it.
}

func (r *adminNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Updating a sent notification is usually not supported.
}

func (r *adminNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data adminNotificationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotification(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification, got error: %s", err))
		return
	}
}

func (r *adminNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
