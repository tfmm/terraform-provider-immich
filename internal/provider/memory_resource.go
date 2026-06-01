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
var _ resource.Resource = &memoryResource{}
var _ resource.ResourceWithImportState = &memoryResource{}

func NewMemoryResource() resource.Resource {
	return &memoryResource{}
}

// memoryResource defines the resource implementation.
type memoryResource struct {
	client *client.Client
}

// memoryResourceModel describes the resource data model.
type memoryResourceModel struct {
	ID       types.String `tfsdk:"id"`
	MemoryAt types.String `tfsdk:"memory_at"`
	IsSaved  types.Bool   `tfsdk:"is_saved"`
}

func (r *memoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_memory"
}

func (r *memoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich memory.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the memory.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"memory_at": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The date the memory represents (ISO 8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_saved": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the memory is saved by the user.",
			},
		},
	}
}

func (r *memoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *memoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data memoryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateMemoryRequest{
		MemoryAt: data.MemoryAt.ValueString(),
		IsSaved:  data.IsSaved.ValueBool(),
	}

	memory, err := r.client.CreateMemory(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create memory, got error: %s", err))
		return
	}

	data.ID = types.StringValue(memory.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *memoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data memoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	memory, err := r.client.GetMemory(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read memory, got error: %s", err))
		return
	}

	data.MemoryAt = types.StringValue(memory.MemoryAt)
	data.IsSaved = types.BoolValue(memory.IsSaved)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *memoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data memoryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	isSaved := data.IsSaved.ValueBool()
	updateReq := client.UpdateMemoryRequest{
		IsSaved: &isSaved,
	}

	_, err := r.client.UpdateMemory(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update memory, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *memoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data memoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMemory(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete memory, got error: %s", err))
		return
	}
}

func (r *memoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
