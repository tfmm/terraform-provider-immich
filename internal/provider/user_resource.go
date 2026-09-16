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
var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}
var _ resource.ResourceWithValidateConfig = &userResource{}

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
	PasswordWO           types.String `tfsdk:"password_wo"`
	PasswordWOVersion    types.Int64  `tfsdk:"password_wo_version"`
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
				Optional:            true,
				Sensitive:           true,
				DeprecationMessage:  "Use `password_wo` instead, which is never persisted to plan or state.",
				MarkdownDescription: "Initial password for the user. Only used during creation or when forced by `should_change_password`. Persisted to state in plain text (aside from standard state encryption); at most one of `password` or `password_wo` may be set. Deprecated in favor of `password_wo`.",
			},
			"password_wo": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				MarkdownDescription: "Write-only initial password for the user. Never persisted to plan or state. Requires Terraform 1.11+. Must be paired with `password_wo_version`; bump the version to rotate the password on a later apply. At most one of `password` or `password_wo` may be set.",
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Arbitrary version number for `password_wo`. Increment it to signal that the password should be rotated; the number itself has no meaning beyond change detection, since `password_wo`'s value is never stored to compare against.",
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

func (r *userResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data userResourceModel
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config userResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := data.Password.ValueString()
	if !config.PasswordWO.IsNull() {
		password = config.PasswordWO.ValueString()
	}

	createReq := client.UserAdminCreateRequest{
		Email:                data.Email.ValueString(),
		Name:                 data.Name.ValueString(),
		Password:             password,
		IsAdmin:              data.IsAdmin.ValueBool(),
		StorageLabel:         data.StorageLabel.ValueString(),
		ShouldChangePassword: data.ShouldChangePassword.ValueBool(),
	}

	if !data.QuotaSizeInBytes.IsNull() {
		quota := data.QuotaSizeInBytes.ValueInt64()
		createReq.QuotaSizeInBytes = &quota
	}

	user, err := r.client.CreateUser(ctx, createReq)
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

	user, err := r.client.GetUser(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	data.Email = types.StringValue(user.Email)
	data.Name = types.StringValue(user.Name)
	data.IsAdmin = types.BoolValue(user.IsAdmin)
	if user.StorageLabel != "" {
		data.StorageLabel = types.StringValue(user.StorageLabel)
	} else {
		data.StorageLabel = types.StringNull()
	}
	if user.QuotaSizeInBytes != nil {
		data.QuotaSizeInBytes = types.Int64Value(*user.QuotaSizeInBytes)
	} else {
		data.QuotaSizeInBytes = types.Int64Null()
	}
	data.ShouldChangePassword = types.BoolValue(user.ShouldChangePassword)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UserAdminUpdateRequest{
		Email:                plan.Email.ValueString(),
		Name:                 plan.Name.ValueString(),
		IsAdmin:              plan.IsAdmin.ValueBool(),
		StorageLabel:         plan.StorageLabel.ValueString(),
		ShouldChangePassword: plan.ShouldChangePassword.ValueBool(),
	}

	switch {
	case !config.PasswordWO.IsNull():
		// The write-only value itself is never available to compare against
		// a prior apply, so a bump in password_wo_version is the only
		// signal that the password should be rotated.
		if !plan.PasswordWOVersion.Equal(state.PasswordWOVersion) {
			updateReq.Password = config.PasswordWO.ValueString()
		}
	case !plan.Password.IsNull() && !plan.Password.Equal(state.Password):
		updateReq.Password = plan.Password.ValueString()
	}

	if !plan.QuotaSizeInBytes.IsNull() {
		quota := plan.QuotaSizeInBytes.ValueInt64()
		updateReq.QuotaSizeInBytes = &quota
	}

	_, err := r.client.UpdateUser(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
