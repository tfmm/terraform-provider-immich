package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/immich-app/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &usersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

// usersDataSource defines the data source implementation.
type usersDataSource struct {
	client *client.Client
}

// usersDataSourceModel describes the data source data model.
type usersDataSourceModel struct {
	Users []usersModel `tfsdk:"users"`
}

type usersModel struct {
	ID           types.String `tfsdk:"id"`
	Email        types.String `tfsdk:"email"`
	Name         types.String `tfsdk:"name"`
	IsAdmin      types.Bool   `tfsdk:"is_admin"`
	StorageLabel types.String `tfsdk:"storage_label"`
}

func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of all Immich users.",

		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of users.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier for the user.",
						},
						"email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Email address of the user.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Full name of the user.",
						},
						"is_admin": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the user has administrative privileges.",
						},
						"storage_label": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Label used for the user's storage path.",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data usersDataSourceModel

	users, err := d.client.GetUsers()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
		return
	}

	for _, user := range users {
		userState := usersModel{
			ID:           types.StringValue(user.ID),
			Email:        types.StringValue(user.Email),
			Name:         types.StringValue(user.Name),
			IsAdmin:      types.BoolValue(user.IsAdmin),
			StorageLabel: types.StringPointerValue(&user.StorageLabel),
		}
		data.Users = append(data.Users, userState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
