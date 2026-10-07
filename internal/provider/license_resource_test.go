// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAidboxLicenseResource(t *testing.T) {
	name := fmt.Sprintf("provider-acc-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAidboxLicenseResourceConfig(name, "development"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aidbox_license.test", "name", name),
					resource.TestCheckResourceAttr("aidbox_license.test", "type", "development"),
					resource.TestCheckResourceAttrSet("aidbox_license.test", "id"),
				),
			},
			{
				ResourceName:      "aidbox_license.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAidboxLicenseResourceConfig(name string, licenseType string) string {
	providerConfig := "provider \"aidbox\" {}"
	if endpoint := os.Getenv("AIDBOX_ENDPOINT"); endpoint != "" {
		providerConfig = fmt.Sprintf("provider \"aidbox\" {\n  endpoint = %q\n}", endpoint)
	}

	return fmt.Sprintf(`
%[3]s

resource "aidbox_license" "test" {
  name = %[1]q
  type = %[2]q
}
`, name, licenseType, providerConfig)
}
