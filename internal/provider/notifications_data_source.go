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
var _ datasource.DataSource = &notificationsDataSource{}

func NewNotificationsDataSource() datasource.DataSource {
	return &notificationsDataSource{}
}

// notificationsDataSource defines the data source implementation.
type notificationsDataSource struct {
	client *client.Client
}

// notificationsDataSourceModel describes the data source data model.
type notificationsDataSourceModel struct {
	UnreadOnly    types.Bool           `tfsdk:"unread_only"`
	Notifications []notificationsModel_ds `tfsdk:"notifications"`
}

type notificationsModel_ds struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	Level       types.String `tfsdk:"level"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *notificationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifications"
}

func (d *notificationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of notifications for the current user.",

		Attributes: map[string]schema.Attribute{
			"unread_only": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Filter for only unread notifications.",
			},
			"notifications": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of notifications.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the notification.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of notification.",
						},
						"level": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Severity level.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Notification title.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Notification message body.",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp when the notification was created.",
						},
					},
				},
			},
		},
	}
}

func (d *notificationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *notificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data notificationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	unread := false
	if !data.UnreadOnly.IsNull() {
		unread = data.UnreadOnly.ValueBool()
	}

	notifications, err := d.client.GetNotifications(unread)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notifications, got error: %s", err))
		return
	}

	data.Notifications = []notificationsModel_ds{}
	for _, n := range notifications {
		nState := notificationsModel_ds{
			ID:          types.StringValue(n.ID),
			Type:        types.StringValue(n.Type),
			Level:       types.StringValue(n.Level),
			Title:       types.StringValue(n.Title),
			Description: types.StringValue(n.Description),
			CreatedAt:   types.StringValue(n.CreatedAt),
		}
		data.Notifications = append(data.Notifications, nState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
