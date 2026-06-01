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
var _ datasource.DataSource = &activitiesDataSource{}

func NewActivitiesDataSource() datasource.DataSource {
	return &activitiesDataSource{}
}

// activitiesDataSource defines the data source implementation.
type activitiesDataSource struct {
	client *client.Client
}

// activitiesDataSourceModel describes the data source data model.
type activitiesDataSourceModel struct {
	AlbumId    types.String      `tfsdk:"album_id"`
	AssetId    types.String      `tfsdk:"asset_id"`
	Activities []activitiesModel `tfsdk:"activities"`
}

type activitiesModel struct {
	ID        types.String `tfsdk:"id"`
	Type      types.String `tfsdk:"type"`
	UserId    types.String `tfsdk:"user_id"`
	Comment   types.String `tfsdk:"comment"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (d *activitiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activities"
}

func (d *activitiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of activities for an album or asset.",

		Attributes: map[string]schema.Attribute{
			"album_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the album.",
			},
			"asset_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the asset.",
			},
			"activities": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of activities.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the activity.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of activity (COMMENT or LIKE).",
						},
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the user who performed the activity.",
						},
						"comment": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Comment text.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp when the activity was created.",
						},
					},
				},
			},
		},
	}
}

func (d *activitiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *activitiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data activitiesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	activities, err := d.client.GetActivities(data.AlbumId.ValueString(), data.AssetId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read activities, got error: %s", err))
		return
	}

	data.Activities = []activitiesModel{}
	for _, activity := range activities {
		activityState := activitiesModel{
			ID:        types.StringValue(activity.ID),
			Type:      types.StringValue(activity.Type),
			UserId:    types.StringValue(activity.User.ID),
			Comment:   types.StringValue(activity.Comment),
			CreatedAt: types.StringValue(activity.CreatedAt),
		}
		data.Activities = append(data.Activities, activityState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
