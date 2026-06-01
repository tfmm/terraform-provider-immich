package provider

import (
	"context"
	"encoding/json"
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
var _ resource.Resource = &workflowResource{}
var _ resource.ResourceWithImportState = &workflowResource{}

func NewWorkflowResource() resource.Resource {
	return &workflowResource{}
}

// workflowResource defines the resource implementation.
type workflowResource struct {
	client *client.Client
}

// workflowResourceModel describes the resource data model.
type workflowResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Triggers types.String `tfsdk:"triggers"` // JSON string for now
	Filters  types.String `tfsdk:"filters"`  // JSON string for now
	Actions  types.String `tfsdk:"actions"`  // JSON string for now
}

func (r *workflowResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *workflowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich workflow (Experimental).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the workflow.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the workflow.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the workflow is enabled.",
			},
			"triggers": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON string representing the workflow triggers.",
			},
			"filters": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "JSON string representing the workflow filters.",
			},
			"actions": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JSON string representing the workflow actions.",
			},
		},
	}
}

func (r *workflowResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data workflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var triggers []map[string]interface{}
	if err := json.Unmarshal([]byte(data.Triggers.ValueString()), &triggers); err != nil {
		resp.Diagnostics.AddError("Invalid Triggers JSON", err.Error())
		return
	}

	var actions []map[string]interface{}
	if err := json.Unmarshal([]byte(data.Actions.ValueString()), &actions); err != nil {
		resp.Diagnostics.AddError("Invalid Actions JSON", err.Error())
		return
	}

	createReq := client.CreateWorkflowRequest{
		Name:     data.Name.ValueString(),
		Enabled:  data.Enabled.ValueBool(),
		Triggers: triggers,
		Actions:  actions,
	}

	if !data.Filters.IsNull() && data.Filters.ValueString() != "" && data.Filters.ValueString() != "false" {
		var filters []map[string]interface{}
		if err := json.Unmarshal([]byte(data.Filters.ValueString()), &filters); err != nil {
			resp.Diagnostics.AddError("Invalid Filters JSON", err.Error())
			return
		}
		createReq.Filters = filters
	}

	workflow, err := r.client.CreateWorkflow(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create workflow, got error: %s", err))
		return
	}

	data.ID = types.StringValue(workflow.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data workflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workflow, err := r.client.GetWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read workflow, got error: %s", err))
		return
	}

	data.Name = types.StringValue(workflow.Name)
	data.Enabled = types.BoolValue(workflow.Enabled)

	triggersJSON, _ := json.Marshal(workflow.Triggers)
	data.Triggers = types.StringValue(string(triggersJSON))

	filtersJSON, _ := json.Marshal(workflow.Filters)
	data.Filters = types.StringValue(string(filtersJSON))

	actionsJSON, _ := json.Marshal(workflow.Actions)
	data.Actions = types.StringValue(string(actionsJSON))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data workflowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var triggers []map[string]interface{}
	if err := json.Unmarshal([]byte(data.Triggers.ValueString()), &triggers); err != nil {
		resp.Diagnostics.AddError("Invalid Triggers JSON", err.Error())
		return
	}

	var actions []map[string]interface{}
	if err := json.Unmarshal([]byte(data.Actions.ValueString()), &actions); err != nil {
		resp.Diagnostics.AddError("Invalid Actions JSON", err.Error())
		return
	}

	enabled := data.Enabled.ValueBool()
	updateReq := client.UpdateWorkflowRequest{
		Name:     data.Name.ValueString(),
		Enabled:  &enabled,
		Triggers: triggers,
		Actions:  actions,
	}

	if !data.Filters.IsNull() && data.Filters.ValueString() != "" && data.Filters.ValueString() != "false" {
		var filters []map[string]interface{}
		if err := json.Unmarshal([]byte(data.Filters.ValueString()), &filters); err != nil {
			resp.Diagnostics.AddError("Invalid Filters JSON", err.Error())
			return
		}
		updateReq.Filters = filters
	}

	_, err := r.client.UpdateWorkflow(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update workflow, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data workflowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete workflow, got error: %s", err))
		return
	}
}

func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
