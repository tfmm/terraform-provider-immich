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
var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource defines the resource implementation.
type userResource struct {
	client *client.Client
}

// userResourceModel describes the resource data model.
type userResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Email                types.String `tfsdk:"email"`
	Name                 types.String `tfsdk:"name"`
	Password             types.String `tfsdk:"password"`
	IsAdmin              types.Bool   `tfsdk:"is_admin"`
	StorageLabel         types.String `tfsdk:"storage_label"`
	QuotaSizeInBytes     types.Int64  `tfsdk:"quota_size_in_bytes"`
	ShouldChangePassword types.Bool   `tfsdk:"should_change_password"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich user account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email address of the user. This is used for login.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Full name of the user.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial password for the user. Only used during creation or when forced by `should_change_password`.",
			},
			"is_admin": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the user has administrative privileges.",
			},
			"storage_label": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Label used for the user's storage path.",
			},
			"quota_size_in_bytes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum storage quota for the user in bytes. Set to 0 or null for unlimited.",
			},
			"should_change_password": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force the user to change their password on next login.",
			},
		},
	}
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.UserAdminCreateRequest{
		Email:                data.Email.ValueString(),
		Name:                 data.Name.ValueString(),
		Password:             data.Password.ValueString(),
		IsAdmin:              data.IsAdmin.ValueBool(),
		StorageLabel:         data.StorageLabel.ValueString(),
		ShouldChangePassword: data.ShouldChangePassword.ValueBool(),
	}

	if !data.QuotaSizeInBytes.IsNull() {
		quota := data.QuotaSizeInBytes.ValueInt64()
		createReq.QuotaSizeInBytes = &quota
	}

	user, err := r.client.CreateUser(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	data.ID = types.StringValue(user.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.Email = types.StringValue(user.Email)
	data.Name = types.StringValue(user.Name)
	data.IsAdmin = types.BoolValue(user.IsAdmin)
	data.StorageLabel = types.StringPointerValue(&user.StorageLabel)
	if user.QuotaSizeInBytes != nil {
		data.QuotaSizeInBytes = types.Int64Value(*user.QuotaSizeInBytes)
	} else {
		data.QuotaSizeInBytes = types.Int64Null()
	}
	data.ShouldChangePassword = types.BoolValue(user.ShouldChangePassword)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UserAdminUpdateRequest{
		Email:                data.Email.ValueString(),
		Name:                 data.Name.ValueString(),
		IsAdmin:              data.IsAdmin.ValueBool(),
		StorageLabel:         data.StorageLabel.ValueString(),
		ShouldChangePassword: data.ShouldChangePassword.ValueBool(),
	}

	if !data.Password.IsNull() {
		updateReq.Password = data.Password.ValueString()
	}

	if !data.QuotaSizeInBytes.IsNull() {
		quota := data.QuotaSizeInBytes.ValueInt64()
		updateReq.QuotaSizeInBytes = &quota
	}

	_, err := r.client.UpdateUser(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
