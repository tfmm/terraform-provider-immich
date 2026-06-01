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
var _ resource.Resource = &faceResource{}
var _ resource.ResourceWithImportState = &faceResource{}

func NewFaceResource() resource.Resource {
	return &faceResource{}
}

// faceResource defines the resource implementation.
type faceResource struct {
	client *client.Client
}

// faceResourceModel describes the resource data model.
type faceResourceModel struct {
	ID            types.String  `tfsdk:"id"`
	AssetId       types.String  `tfsdk:"asset_id"`
	PersonId      types.String  `tfsdk:"person_id"`
	BoundingBoxX1 types.Float64 `tfsdk:"bounding_box_x1"`
	BoundingBoxY1 types.Float64 `tfsdk:"bounding_box_y1"`
	BoundingBoxX2 types.Float64 `tfsdk:"bounding_box_x2"`
	BoundingBoxY2 types.Float64 `tfsdk:"bounding_box_y2"`
	ImageHeight   types.Int64   `tfsdk:"image_height"`
	ImageWidth    types.Int64   `tfsdk:"image_width"`
}

func (r *faceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_face"
}

func (r *faceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Immich face (detected or manual).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the face.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"asset_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the asset this face belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"person_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the person associated with this face.",
			},
			"bounding_box_x1": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The left coordinate of the bounding box.",
				PlanModifiers: []planmodifier.Float64{
					// RequiresReplace because we can't update bounding box via PUT /faces
				},
			},
			"bounding_box_y1": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The top coordinate of the bounding box.",
			},
			"bounding_box_x2": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The right coordinate of the bounding box.",
			},
			"bounding_box_y2": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "The bottom coordinate of the bounding box.",
			},
			"image_height": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The height of the image in pixels.",
			},
			"image_width": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The width of the image in pixels.",
			},
		},
	}
}

func (r *faceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *faceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data faceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateFaceRequest{
		AssetId:       data.AssetId.ValueString(),
		PersonId:      data.PersonId.ValueString(),
		BoundingBoxX1: data.BoundingBoxX1.ValueFloat64(),
		BoundingBoxY1: data.BoundingBoxY1.ValueFloat64(),
		BoundingBoxX2: data.BoundingBoxX2.ValueFloat64(),
		BoundingBoxY2: data.BoundingBoxY2.ValueFloat64(),
		ImageHeight:   int(data.ImageHeight.ValueInt64()),
		ImageWidth:    int(data.ImageWidth.ValueInt64()),
	}

	face, err := r.client.CreateFace(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create face, got error: %s", err))
		return
	}

	data.ID = types.StringValue(face.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *faceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data faceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// We have to find the face in the list for the asset
	faces, err := r.client.GetFaces(data.AssetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read faces for asset, got error: %s", err))
		return
	}

	var found *client.Face
	for _, f := range faces {
		if f.ID == data.ID.ValueString() {
			found = &f
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.PersonId = types.StringValue(found.PersonId)
	data.BoundingBoxX1 = types.Float64Value(found.BoundingBoxX1)
	data.BoundingBoxY1 = types.Float64Value(found.BoundingBoxY1)
	data.BoundingBoxX2 = types.Float64Value(found.BoundingBoxX2)
	data.BoundingBoxY2 = types.Float64Value(found.BoundingBoxY2)
	data.ImageHeight = types.Int64Value(int64(found.ImageHeight))
	data.ImageWidth = types.Int64Value(int64(found.ImageWidth))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *faceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data faceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateFaceRequest{
		PersonId: data.PersonId.ValueString(),
	}

	_, err := r.client.UpdateFace(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update face, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *faceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data faceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFace(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete face, got error: %s", err))
		return
	}
}

func (r *faceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected format: face_id
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
