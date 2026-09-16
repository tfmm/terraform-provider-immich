package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func TestProviderInitialization(t *testing.T) {
	pFunc := New("1.0.0-test")
	p := pFunc()

	ctx := context.Background()

	metaReq := provider.MetadataRequest{}
	metaResp := &provider.MetadataResponse{}
	p.Metadata(ctx, metaReq, metaResp)

	if metaResp.TypeName != "immich" {
		t.Errorf("expected TypeName 'immich', got %q", metaResp.TypeName)
	}
	if metaResp.Version != "1.0.0-test" {
		t.Errorf("expected Version '1.0.0-test', got %q", metaResp.Version)
	}

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, schemaReq, schemaResp)

	if _, ok := schemaResp.Schema.Attributes["endpoint"]; !ok {
		t.Errorf("expected endpoint attribute in provider schema")
	}
	if _, ok := schemaResp.Schema.Attributes["api_key"]; !ok {
		t.Errorf("expected api_key attribute in provider schema")
	}

	immichP, ok := p.(*immichProvider)
	if !ok {
		t.Fatalf("expected *immichProvider")
	}

	resources := immichP.Resources(ctx)
	if len(resources) != 16 {
		t.Errorf("expected 16 registered resources, got %d", len(resources))
	}

	dataSources := immichP.DataSources(ctx)
	if len(dataSources) != 8 {
		t.Errorf("expected 8 registered data sources, got %d", len(dataSources))
	}
}

func TestProviderSchemaAttributes(t *testing.T) {
	pFunc := New("test")()
	resp := &provider.SchemaResponse{}
	pFunc.Schema(context.Background(), provider.SchemaRequest{}, resp)
	s := resp.Schema

	endpointAttr := s.Attributes["endpoint"].(schema.StringAttribute)
	if !endpointAttr.Optional {
		t.Errorf("expected endpoint attribute to be optional")
	}

	apiKeyAttr := s.Attributes["api_key"].(schema.StringAttribute)
	if !apiKeyAttr.Sensitive {
		t.Errorf("expected api_key attribute to be sensitive")
	}
}
