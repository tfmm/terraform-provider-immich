package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &facesDataSource{}

func NewFacesDataSource() datasource.DataSource {
	return &facesDataSource{}
}

// facesDataSource defines the data source implementation.
type facesDataSource struct {
	client *client.Client
}

// facesDataSourceModel describes the data source data model.
type facesDataSourceModel struct {
	AssetId types.String `tfsdk:"asset_id"`
	Faces   []facesModel `tfsdk:"faces"`
}

type facesModel struct {
	ID            types.String  `tfsdk:"id"`
	PersonId      types.String  `tfsdk:"person_id"`
	BoundingBoxX1 types.Float64 `tfsdk:"bounding_box_x1"`
	BoundingBoxY1 types.Float64 `tfsdk:"bounding_box_y1"`
	BoundingBoxX2 types.Float64 `tfsdk:"bounding_box_x2"`
	BoundingBoxY2 types.Float64 `tfsdk:"bounding_box_y2"`
	ImageHeight   types.Int64   `tfsdk:"image_height"`
	ImageWidth    types.Int64   `tfsdk:"image_width"`
}

func (d *facesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_faces"
}

func (d *facesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of faces for a specific asset.",

		Attributes: map[string]schema.Attribute{
			"asset_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the asset.",
			},
			"faces": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of faces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the face.",
						},
						"person_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the person associated with this face.",
						},
						"bounding_box_x1": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The left coordinate of the bounding box.",
						},
						"bounding_box_y1": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The top coordinate of the bounding box.",
						},
						"bounding_box_x2": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The right coordinate of the bounding box.",
						},
						"bounding_box_y2": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "The bottom coordinate of the bounding box.",
						},
						"image_height": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The height of the image in pixels.",
						},
						"image_width": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The width of the image in pixels.",
						},
					},
				},
			},
		},
	}
}

func (d *facesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *facesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data facesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	faces, err := d.client.GetFaces(data.AssetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read faces, got error: %s", err))
		return
	}

	data.Faces = []facesModel{}
	for _, f := range faces {
		fState := facesModel{
			ID:            types.StringValue(f.ID),
			PersonId:      types.StringValue(f.PersonId),
			BoundingBoxX1: types.Float64Value(f.BoundingBoxX1),
			BoundingBoxY1: types.Float64Value(f.BoundingBoxY1),
			BoundingBoxX2: types.Float64Value(f.BoundingBoxX2),
			BoundingBoxY2: types.Float64Value(f.BoundingBoxY2),
			ImageHeight:   types.Int64Value(int64(f.ImageHeight)),
			ImageWidth:    types.Int64Value(int64(f.ImageWidth)),
		}
		data.Faces = append(data.Faces, fState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
