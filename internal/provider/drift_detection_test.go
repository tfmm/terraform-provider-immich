package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// TestResourceReadRemovesResourceOnNotFound is a regression test: several
// resources' Read() previously surfaced any GetXxx error (including 404s)
// as a hard error instead of calling State.RemoveResource, so a resource
// deleted outside Terraform would break every subsequent plan/apply forever
// instead of being recreated.
func TestResourceReadRemovesResourceOnNotFound(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() resource.Resource
		newModel    func(id string) interface{}
	}{
		{"album", func() resource.Resource { return NewAlbumResource() }, func(id string) interface{} { return &albumResourceModel{ID: types.StringValue(id)} }},
		{"apiKey", func() resource.Resource { return NewApiKeyResource() }, func(id string) interface{} { return &apiKeyResourceModel{ID: types.StringValue(id)} }},
		{"asset", func() resource.Resource { return NewAssetResource() }, func(id string) interface{} { return &assetResourceModel{ID: types.StringValue(id)} }},
		{"library", func() resource.Resource { return NewLibraryResource() }, func(id string) interface{} { return &libraryResourceModel{ID: types.StringValue(id)} }},
		{"memory", func() resource.Resource { return NewMemoryResource() }, func(id string) interface{} { return &memoryResourceModel{ID: types.StringValue(id)} }},
		{"person", func() resource.Resource { return NewPersonResource() }, func(id string) interface{} { return &personResourceModel{ID: types.StringValue(id)} }},
		{"sharedLink", func() resource.Resource { return NewSharedLinkResource() }, func(id string) interface{} { return &sharedLinkResourceModel{ID: types.StringValue(id)} }},
		{"stack", func() resource.Resource { return NewStackResource() }, func(id string) interface{} { return &stackResourceModel{ID: types.StringValue(id)} }},
		{"tag", func() resource.Resource { return NewTagResource() }, func(id string) interface{} { return &tagResourceModel{ID: types.StringValue(id)} }},
		{"user", func() resource.Resource { return NewUserResource() }, func(id string) interface{} { return &userResourceModel{ID: types.StringValue(id)} }},
		{"workflow", func() resource.Resource { return NewWorkflowResource() }, func(id string) interface{} { return &workflowResourceModel{ID: types.StringValue(id)} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message":"Not Found"}`))
			}))
			defer server.Close()

			ctx := context.Background()
			c := client.NewClient(server.URL, "test-key")
			res := tt.constructor()

			configurable, ok := res.(resource.ResourceWithConfigure)
			if !ok {
				t.Fatalf("%s: resource does not implement ResourceWithConfigure", tt.name)
			}
			configurable.Configure(ctx, resource.ConfigureRequest{ProviderData: c}, &resource.ConfigureResponse{})

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("%s: unexpected schema diagnostics: %v", tt.name, schemaResp.Diagnostics)
			}

			var state tfsdk.State
			state.Schema = schemaResp.Schema
			if diags := state.Set(ctx, tt.newModel("does-not-exist")); diags.HasError() {
				t.Fatalf("%s: unexpected diagnostics setting state: %v", tt.name, diags)
			}

			readResp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			res.Read(ctx, resource.ReadRequest{State: state}, readResp)

			if readResp.Diagnostics.HasError() {
				t.Fatalf("%s: expected no error diagnostics on 404, got: %v", tt.name, readResp.Diagnostics)
			}
			if !readResp.State.Raw.IsNull() {
				t.Errorf("%s: expected resource to be removed from state on 404, but it was not", tt.name)
			}
		})
	}
}
