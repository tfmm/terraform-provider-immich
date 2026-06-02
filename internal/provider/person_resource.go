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
var _ resource.Resource = &personResource{}
var _ resource.ResourceWithImportState = &personResource{}

func NewPersonResource() resource.Resource {
	return &personResource{}
}

// personResource defines the resource implementation.
type personResource struct {
	client *client.Client
}

// personResourceModel describes the resource data model.
type personResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	BirthDate  types.String `tfsdk:"birth_date"`
	IsHidden   types.Bool   `tfsdk:"is_hidden"`
	IsFavorite types.Bool   `tfsdk:"is_favorite"`
}

func (r *personResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

func (r *personResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich person. Note: Persons are usually created automatically by Immich facial recognition. This resource can be used to update their details or manually create a new person.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Unique identifier for the person.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the person.",
			},
			"birth_date": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Birth date of the person (ISO 8601).",
			},
			"is_hidden": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the person is hidden from the UI.",
			},
			"is_favorite": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the person is marked as a favorite.",
			},
		},
	}
}

func (r *personResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *personResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data personResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, we treat it as an Update (to support managing existing without import if needed, 
	// though not idiomatic)
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		isHidden := data.IsHidden.ValueBool()
		isFavorite := data.IsFavorite.ValueBool()

		updateReq := client.UpdatePersonRequest{
			Name:       data.Name.ValueString(),
			BirthDate:  data.BirthDate.ValueString(),
			IsHidden:   &isHidden,
			IsFavorite: &isFavorite,
		}

		person, err := r.client.UpdatePerson(data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update person, got error: %s", err))
			return
		}

		data.Name = types.StringValue(person.Name)
		data.BirthDate = types.StringValue(person.BirthDate)
		data.IsHidden = types.BoolValue(person.IsHidden)
		data.IsFavorite = types.BoolValue(person.IsFavorite)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Otherwise, create a new person
	createReq := client.CreatePersonRequest{
		Name:       data.Name.ValueString(),
		BirthDate:  data.BirthDate.ValueString(),
		IsHidden:   data.IsHidden.ValueBool(),
		IsFavorite: data.IsFavorite.ValueBool(),
	}

	person, err := r.client.CreatePerson(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create person, got error: %s", err))
		return
	}

	data.ID = types.StringValue(person.ID)
	data.Name = types.StringValue(person.Name)
	data.BirthDate = types.StringValue(person.BirthDate)
	data.IsHidden = types.BoolValue(person.IsHidden)
	data.IsFavorite = types.BoolValue(person.IsFavorite)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *personResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data personResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	person, err := r.client.GetPerson(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read person, got error: %s", err))
		return
	}

	data.Name = types.StringValue(person.Name)
	data.BirthDate = types.StringValue(person.BirthDate)
	data.IsHidden = types.BoolValue(person.IsHidden)
	data.IsFavorite = types.BoolValue(person.IsFavorite)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *personResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data personResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	isHidden := data.IsHidden.ValueBool()
	isFavorite := data.IsFavorite.ValueBool()

	updateReq := client.UpdatePersonRequest{
		Name:       data.Name.ValueString(),
		BirthDate:  data.BirthDate.ValueString(),
		IsHidden:   &isHidden,
		IsFavorite: &isFavorite,
	}

	person, err := r.client.UpdatePerson(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update person, got error: %s", err))
		return
	}

	data.Name = types.StringValue(person.Name)
	data.BirthDate = types.StringValue(person.BirthDate)
	data.IsHidden = types.BoolValue(person.IsHidden)
	data.IsFavorite = types.BoolValue(person.IsFavorite)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *personResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data personResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePerson(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete person, got error: %s", err))
		return
	}
}

func (r *personResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
