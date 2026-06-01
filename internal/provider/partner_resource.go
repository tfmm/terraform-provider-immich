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
var _ resource.Resource = &partnerResource{}
var _ resource.ResourceWithImportState = &partnerResource{}

func NewPartnerResource() resource.Resource {
	return &partnerResource{}
}

// partnerResource defines the resource implementation.
type partnerResource struct {
	client *client.Client
}

// partnerResourceModel describes the resource data model.
type partnerResourceModel struct {
	PartnerId  types.String `tfsdk:"partner_id"`
	InTimeline types.Bool   `tfsdk:"in_timeline"`
	Email      types.String `tfsdk:"email"`
	Name       types.String `tfsdk:"name"`
}

func (r *partnerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner"
}

func (r *partnerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich partner connection.",

		Attributes: map[string]schema.Attribute{
			"partner_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the user to partner with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"in_timeline": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the partner's assets should appear in your timeline.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Email of the partner user.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the partner user.",
			},
		},
	}
}

func (r *partnerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *partnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data partnerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	partner, err := r.client.CreatePartner(data.PartnerId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create partner, got error: %s", err))
		return
	}

	// Update timeline visibility if different from default
	if !data.InTimeline.IsNull() && data.InTimeline.ValueBool() == false {
		_, err = r.client.UpdatePartner(data.PartnerId.ValueString(), client.UpdatePartnerRequest{
			InTimeline: false,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update partner timeline visibility, got error: %s", err))
			return
		}
	}

	data.Email = types.StringValue(partner.Email)
	data.Name = types.StringValue(partner.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *partnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data partnerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	partners, err := r.client.GetPartners()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read partners, got error: %s", err))
		return
	}

	var found *client.Partner
	for _, p := range partners {
		if p.ID == data.PartnerId.ValueString() {
			found = &p
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Email = types.StringValue(found.Email)
	data.Name = types.StringValue(found.Name)
	data.InTimeline = types.BoolValue(found.InTimeline)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *partnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data partnerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdatePartner(data.PartnerId.ValueString(), client.UpdatePartnerRequest{
		InTimeline: data.InTimeline.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update partner, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *partnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data partnerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePartner(data.PartnerId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete partner, got error: %s", err))
		return
	}
}

func (r *partnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("partner_id"), req, resp)
}
