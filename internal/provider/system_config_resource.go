package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/immich-app/terraform-provider-immich/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &systemConfigResource{}

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
	
	PasswordLogin   *passwordLoginModel   `tfsdk:"password_login"`
	OAuth           *oauthModel           `tfsdk:"oauth"`
	StorageTemplate *storageTemplateModel `tfsdk:"storage_template"`
	MachineLearning *machineLearningModel `tfsdk:"machine_learning"`
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

	_, err = r.client.UpdateSystemConfig(newConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update system config, got error: %s", err))
		return
	}

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

	_, err = r.client.UpdateSystemConfig(newConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update system config, got error: %s", err))
		return
	}

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
		config.PasswordLogin["enabled"] = model.PasswordLogin.Enabled.ValueBool()
	}

	if model.OAuth != nil {
		if config.OAuth == nil {
			config.OAuth = make(map[string]interface{})
		}
		config.OAuth["enabled"] = model.OAuth.Enabled.ValueBool()
		config.OAuth["issuerUrl"] = model.OAuth.IssuerUrl.ValueString()
		config.OAuth["clientId"] = model.OAuth.ClientId.ValueString()
		config.OAuth["clientSecret"] = model.OAuth.ClientSecret.ValueString()
		config.OAuth["scope"] = model.OAuth.Scope.ValueString()
		config.OAuth["buttonText"] = model.OAuth.ButtonText.ValueString()
		config.OAuth["autoLaunch"] = model.OAuth.AutoLaunch.ValueBool()
		config.OAuth["autoRegister"] = model.OAuth.AutoRegister.ValueBool()
		config.OAuth["mobileOverrideUrl"] = model.OAuth.MobileOverrideUrl.ValueString()
		config.OAuth["mobileRedirectUri"] = model.OAuth.MobileRedirectUri.ValueString()
		config.OAuth["signingAlgorithm"] = model.OAuth.SigningAlgorithm.ValueString()
		config.OAuth["defaultStorageQuota"] = model.OAuth.DefaultStorageQuota.ValueInt64()
	}

	if model.StorageTemplate != nil {
		if config.StorageTemplate == nil {
			config.StorageTemplate = make(map[string]interface{})
		}
		config.StorageTemplate["template"] = model.StorageTemplate.Template.ValueString()
	}

	if model.MachineLearning != nil {
		if config.MachineLearning == nil {
			config.MachineLearning = make(map[string]interface{})
		}
		config.MachineLearning["enabled"] = model.MachineLearning.Enabled.ValueBool()
		config.MachineLearning["url"] = model.MachineLearning.URL.ValueString()
		config.MachineLearning["clipModel"] = model.MachineLearning.ClipModel.ValueString()
		config.MachineLearning["facialRecognitionModel"] = model.MachineLearning.FacialRecognitionModel.ValueString()
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
		}
	}

	if config.OAuth != nil {
		if model.OAuth == nil {
			model.OAuth = &oauthModel{}
		}
		if v, ok := config.OAuth["enabled"].(bool); ok {
			model.OAuth.Enabled = types.BoolValue(v)
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
		}
		if v, ok := config.OAuth["autoRegister"].(bool); ok {
			model.OAuth.AutoRegister = types.BoolValue(v)
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

	return model
}
