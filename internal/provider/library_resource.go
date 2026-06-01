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
var _ resource.Resource = &libraryResource{}
var _ resource.ResourceWithImportState = &libraryResource{}

func NewLibraryResource() resource.Resource {
	return &libraryResource{}
}

// libraryResource defines the resource implementation.
type libraryResource struct {
	client *client.Client
}

// libraryResourceModel describes the resource data model.
type libraryResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	OwnerId           types.String   `tfsdk:"owner_id"`
	Name              types.String   `tfsdk:"name"`
	Type              types.String   `tfsdk:"type"`
	ImportPaths       []types.String `tfsdk:"import_paths"`
	ExclusionPatterns []types.String `tfsdk:"exclusion_patterns"`
	IsVisible         types.Bool     `tfsdk:"is_visible"`
	AssetCount        types.Int64    `tfsdk:"asset_count"`
}

func (r *libraryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_library"
}

func (r *libraryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich library.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the library.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the library owner.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the library.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of the library. Must be either `UPLOAD` or `EXTERNAL`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"import_paths": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of filesystem paths to import assets from (required for `EXTERNAL` libraries).",
			},
			"exclusion_patterns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of glob patterns to exclude from import.",
			},
			"is_visible": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the library is visible in the UI.",
			},
			"asset_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of assets currently in the library.",
			},
		},
	}
}

func (r *libraryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *libraryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data libraryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	importPaths := make([]string, len(data.ImportPaths))
	for i, p := range data.ImportPaths {
		importPaths[i] = p.ValueString()
	}

	exclusionPatterns := make([]string, len(data.ExclusionPatterns))
	for i, p := range data.ExclusionPatterns {
		exclusionPatterns[i] = p.ValueString()
	}

	createReq := client.CreateLibraryRequest{
		Name:              data.Name.ValueString(),
		Type:              data.Type.ValueString(),
		ImportPaths:       importPaths,
		ExclusionPatterns: exclusionPatterns,
		IsVisible:         data.IsVisible.ValueBool(),
	}

	library, err := r.client.CreateLibrary(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create library, got error: %s", err))
		return
	}

	data.ID = types.StringValue(library.ID)
	data.OwnerId = types.StringValue(library.OwnerId)
	data.AssetCount = types.Int64Value(int64(library.AssetCount))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *libraryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data libraryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	library, err := r.client.GetLibrary(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read library, got error: %s", err))
		return
	}

	data.Name = types.StringValue(library.Name)
	data.Type = types.StringValue(library.Type)
	data.OwnerId = types.StringValue(library.OwnerId)
	data.AssetCount = types.Int64Value(int64(library.AssetCount))

	importPaths := make([]types.String, len(library.ImportPaths))
	for i, p := range library.ImportPaths {
		importPaths[i] = types.StringValue(p)
	}
	data.ImportPaths = importPaths

	exclusionPatterns := make([]types.String, len(library.ExclusionPatterns))
	for i, p := range library.ExclusionPatterns {
		exclusionPatterns[i] = types.StringValue(p)
	}
	data.ExclusionPatterns = exclusionPatterns

	// Note: isVisible is not returned in LibraryResponseDto, we'll keep what's in state
	// or assume it's true if not available.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *libraryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data libraryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	importPaths := make([]string, len(data.ImportPaths))
	for i, p := range data.ImportPaths {
		importPaths[i] = p.ValueString()
	}

	exclusionPatterns := make([]string, len(data.ExclusionPatterns))
	for i, p := range data.ExclusionPatterns {
		exclusionPatterns[i] = p.ValueString()
	}

	isVisible := data.IsVisible.ValueBool()
	updateReq := client.UpdateLibraryRequest{
		Name:              data.Name.ValueString(),
		ImportPaths:       importPaths,
		ExclusionPatterns: exclusionPatterns,
		IsVisible:         &isVisible,
	}

	_, err := r.client.UpdateLibrary(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update library, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *libraryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data libraryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLibrary(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete library, got error: %s", err))
		return
	}
}

func (r *libraryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
