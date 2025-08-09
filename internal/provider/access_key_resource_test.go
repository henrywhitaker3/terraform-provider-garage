// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccAccessKeyResource(t *testing.T) {
	garage, cancel := garage(t)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: garage + testAccAccessKeyResourceNeverExpiresConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_access_key.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("bongo"),
					),
					statecheck.ExpectSensitiveValue(
						"garage_access_key.test",
						tfjsonpath.New("access_key_id"),
					),
					statecheck.ExpectSensitiveValue(
						"garage_access_key.test",
						tfjsonpath.New("secret_access_key"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:            "garage_access_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_access_key"},
			},
			// Update and Read testing
			{
				Config: garage + testAccAccessKeyResourceNeverExpiresConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_access_key.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("bongo"),
					),
				},
			},
		},
	})
}

func testAccAccessKeyResourceNeverExpiresConfig() string {
	return `
resource "garage_access_key" "test" {
	name = "bongo"
	never_expires = true
}
`
}
