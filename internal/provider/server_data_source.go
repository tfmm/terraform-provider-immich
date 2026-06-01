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
var _ datasource.DataSource = &serverDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// serverDataSource defines the data source implementation.
type serverDataSource struct {
	client *client.Client
}

// serverDataSourceModel describes the data source data model.
type serverDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Version     types.String `tfsdk:"version"`
	Build       types.String `tfsdk:"build"`
	NodeJS      types.String `tfsdk:"nodejs"`
	FFmpeg      types.String `tfsdk:"ffmpeg"`
	ExifTool    types.String `tfsdk:"exiftool"`
	ImageMagick types.String `tfsdk:"imagemagick"`
	Libvips     types.String `tfsdk:"libvips"`
}

func (d *serverDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_info"
}

func (d *serverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about the Immich server.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server version.",
			},
			"build": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Build ID.",
			},
			"nodejs": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Node.js version.",
			},
			"ffmpeg": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "FFmpeg version.",
			},
			"exiftool": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ExifTool version.",
			},
			"imagemagick": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ImageMagick version.",
			},
			"libvips": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "libvips version.",
			},
		},
	}
}

func (d *serverDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverDataSourceModel

	about, err := d.client.GetServerAbout()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server about, got error: %s", err))
		return
	}

	data.ID = types.StringValue("server_info")
	data.Version = types.StringValue(about.Version)
	data.Build = types.StringValue(about.Build)
	data.NodeJS = types.StringValue(about.NodeJS)
	data.FFmpeg = types.StringValue(about.FFmpeg)
	data.ExifTool = types.StringValue(about.ExifTool)
	data.ImageMagick = types.StringValue(about.ImageMagick)
	data.Libvips = types.StringValue(about.Libvips)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
