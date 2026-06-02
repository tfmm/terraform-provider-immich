package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &systemConfigResource{}
var _ resource.ResourceWithImportState = &systemConfigResource{}

func NewSystemConfigResource() resource.Resource {
	return &systemConfigResource{}
}

// systemConfigResource defines the resource implementation.
type systemConfigResource struct {
	client *client.Client
}

// systemConfigResourceModel describes the resource data model.
type systemConfigResourceModel struct {
	// For simplicity in this implementation, we'll use map[string]types.Map or similar if possible.
	// But Terraform plugin framework works best with explicit nested attributes.
	// To keep it manageable and robust, we'll focus on some common sections first.
	
	ID              types.String          `tfsdk:"id"`
	PasswordLogin   *passwordLoginModel   `tfsdk:"password_login"`
	OAuth           *oauthModel           `tfsdk:"oauth"`
	StorageTemplate *storageTemplateModel `tfsdk:"storage_template"`
	MachineLearning *machineLearningModel `tfsdk:"machine_learning"`
	Notifications   *notificationsModel   `tfsdk:"notifications"`
	Templates       *templatesModel       `tfsdk:"templates"`
}

type passwordLoginModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type machineLearningModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	URL            types.String `tfsdk:"url"`
	ClipModel      types.String `tfsdk:"clip_model"`
	FacialRecognitionModel types.String `tfsdk:"facial_recognition_model"`
}

type notificationsModel struct {
	SMTP *smtpModel `tfsdk:"smtp"`
}

type smtpModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	From       types.String `tfsdk:"from"`
	ReplyTo    types.String `tfsdk:"reply_to"`
	Secure     types.Bool   `tfsdk:"secure"`
	IgnoreCert types.Bool   `tfsdk:"ignore_cert"`
}

type templatesModel struct {
	Email *emailTemplatesModel `tfsdk:"email"`
}

type emailTemplatesModel struct {
	AlbumInviteTemplate types.String `tfsdk:"album_invite_template"`
	AlbumUpdateTemplate types.String `tfsdk:"album_update_template"`
	WelcomeTemplate     types.String `tfsdk:"welcome_template"`
}

type oauthModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	IssuerUrl          types.String `tfsdk:"issuer_url"`
	ClientId           types.String `tfsdk:"client_id"`
	ClientSecret       types.String `tfsdk:"client_secret"`
	Scope              types.String `tfsdk:"scope"`
	ButtonText         types.String `tfsdk:"button_text"`
	AutoLaunch         types.Bool   `tfsdk:"auto_launch"`
	AutoRegister       types.Bool   `tfsdk:"auto_register"`
	MobileOverrideUrl  types.String `tfsdk:"mobile_override_url"`
	MobileRedirectUri  types.String `tfsdk:"mobile_redirect_uri"`
	SigningAlgorithm   types.String `tfsdk:"signing_algorithm"`
	DefaultStorageQuota types.Int64  `tfsdk:"default_storage_quota"`
}

type storageTemplateModel struct {
	Template types.String `tfsdk:"template"`
}

func (r *systemConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

func (r *systemConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Immich system configuration. This is a singleton resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Virtual identifier for the singleton resource.",
			},
			"password_login": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Enable password login.",
					},
				},
			},
			"oauth": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Enable OAuth login.",
					},
					"issuer_url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "OAuth issuer URL.",
					},
					"client_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "OAuth client ID.",
					},
					"client_secret": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "OAuth client secret.",
					},
					"scope": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "OAuth scope.",
					},
					"button_text": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "OAuth button text.",
					},
					"auto_launch": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Auto launch OAuth login.",
					},
					"auto_register": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Auto register users via OAuth.",
					},
					"mobile_override_url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Mobile override URL.",
					},
					"mobile_redirect_uri": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Mobile redirect URI.",
					},
					"signing_algorithm": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Signing algorithm.",
					},
					"default_storage_quota": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Default storage quota for new users in bytes.",
					},
				},
			},
			"storage_template": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"template": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Storage template (e.g. `{{y}}/{{y}}-{{m}}-{{d}}/{{filename}}`).",
					},
				},
			},
			"machine_learning": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Enable machine learning features.",
					},
					"url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "URL of the machine learning server.",
					},
					"clip_model": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "CLIP model to use.",
					},
					"facial_recognition_model": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Facial recognition model to use.",
					},
				},
			},
			"notifications": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"smtp": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Enable SMTP email notifications.",
							},
							"host": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "SMTP server hostname.",
							},
							"port": schema.Int64Attribute{
								Optional:            true,
								MarkdownDescription: "SMTP server port.",
							},
							"username": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "SMTP authentication username.",
							},
							"password": schema.StringAttribute{
								Optional:            true,
								Sensitive:           true,
								MarkdownDescription: "SMTP authentication password.",
							},
							"from": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Sender email address.",
							},
							"reply_to": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Reply-to email address.",
							},
							"secure": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether to use TLS/SSL.",
							},
							"ignore_cert": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether to ignore certificate validation errors.",
							},
						},
					},
				},
			},
			"templates": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"email": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"album_invite_template": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Email template for album invitations.",
							},
							"album_update_template": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Email template for album updates.",
							},
							"welcome_template": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Email template for welcome emails.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *systemConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *systemConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data systemConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read existing config first to avoid wiping other sections
	currentConfig, err := r.client.GetSystemConfig()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current system config, got error: %s", err))
		return
	}

	newConfig := r.mapModelToClient(data, *currentConfig)

	updatedConfig, err := r.client.UpdateSystemConfig(newConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update system config, got error: %s", err))
		return
	}

	data = r.mapClientToModel(*updatedConfig, data)
	data.ID = types.StringValue("system_config")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *systemConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data systemConfigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.GetSystemConfig()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read system config, got error: %s", err))
		return
	}

	data = r.mapClientToModel(*config, data)
	data.ID = types.StringValue("system_config")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *systemConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data systemConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read existing config first to avoid wiping other sections
	currentConfig, err := r.client.GetSystemConfig()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current system config, got error: %s", err))
		return
	}

	newConfig := r.mapModelToClient(data, *currentConfig)

	updatedConfig, err := r.client.UpdateSystemConfig(newConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update system config, got error: %s", err))
		return
	}

	data = r.mapClientToModel(*updatedConfig, data)
	data.ID = types.StringValue("system_config")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *systemConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// System config is a singleton and cannot be truly deleted.
	// We could optionally reset to defaults, but for now we just remove from state.
}

func (r *systemConfigResource) mapModelToClient(model systemConfigResourceModel, config client.SystemConfig) client.SystemConfig {
	if model.PasswordLogin != nil {
		if config.PasswordLogin == nil {
			config.PasswordLogin = make(map[string]interface{})
		}
		if !model.PasswordLogin.Enabled.IsNull() && !model.PasswordLogin.Enabled.IsUnknown() {
			config.PasswordLogin["enabled"] = model.PasswordLogin.Enabled.ValueBool()
		}
	}

	if model.OAuth != nil {
		if config.OAuth == nil {
			config.OAuth = make(map[string]interface{})
		}
		if !model.OAuth.Enabled.IsNull() && !model.OAuth.Enabled.IsUnknown() {
			config.OAuth["enabled"] = model.OAuth.Enabled.ValueBool()
		}
		if !model.OAuth.IssuerUrl.IsNull() && !model.OAuth.IssuerUrl.IsUnknown() {
			config.OAuth["issuerUrl"] = model.OAuth.IssuerUrl.ValueString()
		}
		if !model.OAuth.ClientId.IsNull() && !model.OAuth.ClientId.IsUnknown() {
			config.OAuth["clientId"] = model.OAuth.ClientId.ValueString()
		}
		if !model.OAuth.ClientSecret.IsNull() && !model.OAuth.ClientSecret.IsUnknown() {
			config.OAuth["clientSecret"] = model.OAuth.ClientSecret.ValueString()
		}
		if !model.OAuth.Scope.IsNull() && !model.OAuth.Scope.IsUnknown() {
			config.OAuth["scope"] = model.OAuth.Scope.ValueString()
		}
		if !model.OAuth.ButtonText.IsNull() && !model.OAuth.ButtonText.IsUnknown() {
			config.OAuth["buttonText"] = model.OAuth.ButtonText.ValueString()
		}
		if !model.OAuth.AutoLaunch.IsNull() && !model.OAuth.AutoLaunch.IsUnknown() {
			config.OAuth["autoLaunch"] = model.OAuth.AutoLaunch.ValueBool()
		}
		if !model.OAuth.AutoRegister.IsNull() && !model.OAuth.AutoRegister.IsUnknown() {
			config.OAuth["autoRegister"] = model.OAuth.AutoRegister.ValueBool()
		}
		if !model.OAuth.MobileOverrideUrl.IsNull() && !model.OAuth.MobileOverrideUrl.IsUnknown() {
			config.OAuth["mobileOverrideUrl"] = model.OAuth.MobileOverrideUrl.ValueString()
		}
		if !model.OAuth.MobileRedirectUri.IsNull() && !model.OAuth.MobileRedirectUri.IsUnknown() {
			config.OAuth["mobileRedirectUri"] = model.OAuth.MobileRedirectUri.ValueString()
		}
		if !model.OAuth.SigningAlgorithm.IsNull() && !model.OAuth.SigningAlgorithm.IsUnknown() {
			config.OAuth["signingAlgorithm"] = model.OAuth.SigningAlgorithm.ValueString()
		}
		if !model.OAuth.DefaultStorageQuota.IsNull() && !model.OAuth.DefaultStorageQuota.IsUnknown() {
			config.OAuth["defaultStorageQuota"] = model.OAuth.DefaultStorageQuota.ValueInt64()
		}
	}

	if model.StorageTemplate != nil {
		if config.StorageTemplate == nil {
			config.StorageTemplate = make(map[string]interface{})
		}
		if !model.StorageTemplate.Template.IsNull() && !model.StorageTemplate.Template.IsUnknown() {
			config.StorageTemplate["template"] = model.StorageTemplate.Template.ValueString()
		}
	}

	if model.MachineLearning != nil {
		if config.MachineLearning == nil {
			config.MachineLearning = make(map[string]interface{})
		}
		if !model.MachineLearning.Enabled.IsNull() && !model.MachineLearning.Enabled.IsUnknown() {
			config.MachineLearning["enabled"] = model.MachineLearning.Enabled.ValueBool()
		}
		if !model.MachineLearning.URL.IsNull() && !model.MachineLearning.URL.IsUnknown() {
			config.MachineLearning["url"] = model.MachineLearning.URL.ValueString()
		}
		if !model.MachineLearning.ClipModel.IsNull() && !model.MachineLearning.ClipModel.IsUnknown() {
			config.MachineLearning["clipModel"] = model.MachineLearning.ClipModel.ValueString()
		}
		if !model.MachineLearning.FacialRecognitionModel.IsNull() && !model.MachineLearning.FacialRecognitionModel.IsUnknown() {
			config.MachineLearning["facialRecognitionModel"] = model.MachineLearning.FacialRecognitionModel.ValueString()
		}
	}

	if model.Notifications != nil {
		if config.Notifications == nil {
			config.Notifications = make(map[string]interface{})
		}
		if model.Notifications.SMTP != nil {
			smtp := make(map[string]interface{})
			if !model.Notifications.SMTP.Enabled.IsNull() && !model.Notifications.SMTP.Enabled.IsUnknown() {
				smtp["enabled"] = model.Notifications.SMTP.Enabled.ValueBool()
			}
			if !model.Notifications.SMTP.From.IsNull() && !model.Notifications.SMTP.From.IsUnknown() {
				smtp["from"] = model.Notifications.SMTP.From.ValueString()
			}
			if !model.Notifications.SMTP.ReplyTo.IsNull() && !model.Notifications.SMTP.ReplyTo.IsUnknown() {
				smtp["replyTo"] = model.Notifications.SMTP.ReplyTo.ValueString()
			}

			transport := make(map[string]interface{})
			if !model.Notifications.SMTP.Host.IsNull() && !model.Notifications.SMTP.Host.IsUnknown() {
				transport["host"] = model.Notifications.SMTP.Host.ValueString()
			}
			if !model.Notifications.SMTP.Port.IsNull() && !model.Notifications.SMTP.Port.IsUnknown() {
				transport["port"] = model.Notifications.SMTP.Port.ValueInt64()
			}
			if !model.Notifications.SMTP.Username.IsNull() && !model.Notifications.SMTP.Username.IsUnknown() {
				transport["username"] = model.Notifications.SMTP.Username.ValueString()
			}
			if !model.Notifications.SMTP.Password.IsNull() && !model.Notifications.SMTP.Password.IsUnknown() {
				transport["password"] = model.Notifications.SMTP.Password.ValueString()
			}
			if !model.Notifications.SMTP.Secure.IsNull() && !model.Notifications.SMTP.Secure.IsUnknown() {
				transport["secure"] = model.Notifications.SMTP.Secure.ValueBool()
			}
			if !model.Notifications.SMTP.IgnoreCert.IsNull() && !model.Notifications.SMTP.IgnoreCert.IsUnknown() {
				transport["ignoreCert"] = model.Notifications.SMTP.IgnoreCert.ValueBool()
			}

			smtp["transport"] = transport
			config.Notifications["smtp"] = smtp
		}
	}

	if model.Templates != nil {
		if config.Templates == nil {
			config.Templates = make(map[string]interface{})
		}
		if model.Templates.Email != nil {
			email := make(map[string]interface{})
			if !model.Templates.Email.AlbumInviteTemplate.IsNull() && !model.Templates.Email.AlbumInviteTemplate.IsUnknown() {
				email["albumInviteTemplate"] = model.Templates.Email.AlbumInviteTemplate.ValueString()
			}
			if !model.Templates.Email.AlbumUpdateTemplate.IsNull() && !model.Templates.Email.AlbumUpdateTemplate.IsUnknown() {
				email["albumUpdateTemplate"] = model.Templates.Email.AlbumUpdateTemplate.ValueString()
			}
			if !model.Templates.Email.WelcomeTemplate.IsNull() && !model.Templates.Email.WelcomeTemplate.IsUnknown() {
				email["welcomeTemplate"] = model.Templates.Email.WelcomeTemplate.ValueString()
			}
			config.Templates["email"] = email
		}
	}

	return config
}

func (r *systemConfigResource) mapClientToModel(config client.SystemConfig, model systemConfigResourceModel) systemConfigResourceModel {
	if config.PasswordLogin != nil {
		if model.PasswordLogin == nil {
			model.PasswordLogin = &passwordLoginModel{}
		}
		if v, ok := config.PasswordLogin["enabled"].(bool); ok {
			model.PasswordLogin.Enabled = types.BoolValue(v)
		} else if model.PasswordLogin.Enabled.IsUnknown() {
			model.PasswordLogin.Enabled = types.BoolValue(false)
		}
	}

	if config.OAuth != nil {
		if model.OAuth == nil {
			model.OAuth = &oauthModel{}
		}
		if v, ok := config.OAuth["enabled"].(bool); ok {
			model.OAuth.Enabled = types.BoolValue(v)
		} else if model.OAuth.Enabled.IsUnknown() {
			model.OAuth.Enabled = types.BoolValue(false)
		}
		if v, ok := config.OAuth["issuerUrl"].(string); ok {
			model.OAuth.IssuerUrl = types.StringValue(v)
		}
		if v, ok := config.OAuth["clientId"].(string); ok {
			model.OAuth.ClientId = types.StringValue(v)
		}
		// clientSecret is often masked or not returned, we might want to keep it in state if it's sensitive
		if v, ok := config.OAuth["scope"].(string); ok {
			model.OAuth.Scope = types.StringValue(v)
		}
		if v, ok := config.OAuth["buttonText"].(string); ok {
			model.OAuth.ButtonText = types.StringValue(v)
		}
		if v, ok := config.OAuth["autoLaunch"].(bool); ok {
			model.OAuth.AutoLaunch = types.BoolValue(v)
		} else if model.OAuth.AutoLaunch.IsUnknown() {
			model.OAuth.AutoLaunch = types.BoolValue(false)
		}
		if v, ok := config.OAuth["autoRegister"].(bool); ok {
			model.OAuth.AutoRegister = types.BoolValue(v)
		} else if model.OAuth.AutoRegister.IsUnknown() {
			model.OAuth.AutoRegister = types.BoolValue(false)
		}
		if v, ok := config.OAuth["mobileOverrideUrl"].(string); ok {
			model.OAuth.MobileOverrideUrl = types.StringValue(v)
		}
		if v, ok := config.OAuth["mobileRedirectUri"].(string); ok {
			model.OAuth.MobileRedirectUri = types.StringValue(v)
		}
		if v, ok := config.OAuth["signingAlgorithm"].(string); ok {
			model.OAuth.SigningAlgorithm = types.StringValue(v)
		}
		if v, ok := config.OAuth["defaultStorageQuota"].(float64); ok {
			model.OAuth.DefaultStorageQuota = types.Int64Value(int64(v))
		} else if v, ok := config.OAuth["defaultStorageQuota"].(int64); ok {
			model.OAuth.DefaultStorageQuota = types.Int64Value(v)
		}
	}

	if config.StorageTemplate != nil {
		if model.StorageTemplate == nil {
			model.StorageTemplate = &storageTemplateModel{}
		}
		if v, ok := config.StorageTemplate["template"].(string); ok {
			model.StorageTemplate.Template = types.StringValue(v)
		}
	}

	if config.MachineLearning != nil {
		if model.MachineLearning == nil {
			model.MachineLearning = &machineLearningModel{}
		}
		if v, ok := config.MachineLearning["enabled"].(bool); ok {
			model.MachineLearning.Enabled = types.BoolValue(v)
		} else if model.MachineLearning.Enabled.IsUnknown() {
			model.MachineLearning.Enabled = types.BoolValue(false)
		}
		if v, ok := config.MachineLearning["url"].(string); ok {
			model.MachineLearning.URL = types.StringValue(v)
		}
		if v, ok := config.MachineLearning["clipModel"].(string); ok {
			model.MachineLearning.ClipModel = types.StringValue(v)
		}
		if v, ok := config.MachineLearning["facialRecognitionModel"].(string); ok {
			model.MachineLearning.FacialRecognitionModel = types.StringValue(v)
		}
	}

	if config.Notifications != nil {
		if model.Notifications == nil {
			model.Notifications = &notificationsModel{}
		}
		if smtpConfig, ok := config.Notifications["smtp"].(map[string]interface{}); ok {
			if model.Notifications.SMTP == nil {
				model.Notifications.SMTP = &smtpModel{}
			}
			if v, ok := smtpConfig["enabled"].(bool); ok {
				model.Notifications.SMTP.Enabled = types.BoolValue(v)
			} else if model.Notifications.SMTP.Enabled.IsUnknown() {
				model.Notifications.SMTP.Enabled = types.BoolValue(false)
			}
			if v, ok := smtpConfig["from"].(string); ok {
				model.Notifications.SMTP.From = types.StringValue(v)
			}
			if v, ok := smtpConfig["replyTo"].(string); ok {
				model.Notifications.SMTP.ReplyTo = types.StringValue(v)
			}
			if transport, ok := smtpConfig["transport"].(map[string]interface{}); ok {
				if v, ok := transport["host"].(string); ok {
					model.Notifications.SMTP.Host = types.StringValue(v)
				}
				if v, ok := transport["port"].(float64); ok {
					model.Notifications.SMTP.Port = types.Int64Value(int64(v))
				} else if v, ok := transport["port"].(int64); ok {
					model.Notifications.SMTP.Port = types.Int64Value(v)
				}
				if v, ok := transport["username"].(string); ok {
					model.Notifications.SMTP.Username = types.StringValue(v)
				}
				// password usually not returned or masked
				if v, ok := transport["secure"].(bool); ok {
					model.Notifications.SMTP.Secure = types.BoolValue(v)
				} else if model.Notifications.SMTP.Secure.IsUnknown() {
					model.Notifications.SMTP.Secure = types.BoolValue(false)
				}
				if v, ok := transport["ignoreCert"].(bool); ok {
					model.Notifications.SMTP.IgnoreCert = types.BoolValue(v)
				} else if model.Notifications.SMTP.IgnoreCert.IsUnknown() {
					model.Notifications.SMTP.IgnoreCert = types.BoolValue(false)
				}
			}
		}
	}

	if config.Templates != nil {
		if model.Templates == nil {
			model.Templates = &templatesModel{}
		}
		if emailConfig, ok := config.Templates["email"].(map[string]interface{}); ok {
			if model.Templates.Email == nil {
				model.Templates.Email = &emailTemplatesModel{}
			}
			if v, ok := emailConfig["albumInviteTemplate"].(string); ok {
				model.Templates.Email.AlbumInviteTemplate = types.StringValue(v)
			}
			if v, ok := emailConfig["albumUpdateTemplate"].(string); ok {
				model.Templates.Email.AlbumUpdateTemplate = types.StringValue(v)
			}
			if v, ok := emailConfig["welcomeTemplate"].(string); ok {
				model.Templates.Email.WelcomeTemplate = types.StringValue(v)
			}
		}
	}

	return model
}

func (r *systemConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
