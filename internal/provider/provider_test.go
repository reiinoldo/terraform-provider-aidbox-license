// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"aidbox": providerserver.NewProtocol6WithError(New("test")()),
}

func TestCredentialAndLicenseTokenAreSensitive(t *testing.T) {
	var providerResp provider.SchemaResponse
	(&AidboxProvider{}).Schema(context.Background(), provider.SchemaRequest{}, &providerResp)
	token, ok := providerResp.Schema.Attributes["token"].(providerschema.StringAttribute)
	if !ok || !token.Sensitive {
		t.Fatal("provider token must be marked sensitive")
	}

	var resourceResp resource.SchemaResponse
	(&LicenseResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resourceResp)
	jwt, ok := resourceResp.Schema.Attributes["jwt"].(resourceschema.StringAttribute)
	if !ok || !jwt.Sensitive {
		t.Fatal("license JWT must be marked sensitive")
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AIDBOX_API_TOKEN"); v == "" {
		t.Fatal("AIDBOX_API_TOKEN must be set for acceptance tests")
	}
}
