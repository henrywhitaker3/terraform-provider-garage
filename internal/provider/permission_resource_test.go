// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPermissionResource(t *testing.T) {
	garage, cancel := garage(t)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: garage + testAccPermissionResourceConfig(false, false, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("owner"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("read"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("write"),
						knownvalue.Bool(false),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "garage_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: garage + testAccPermissionResourceConfig(false, true, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("owner"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("read"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"garage_permission.test",
						tfjsonpath.New("write"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccPermissionResourceConfig(owner, read, write bool) string {
	return fmt.Sprintf(`
resource "garage_bucket" "test" {
	name = "bongo"
}
resource "garage_access_key" "test" {
	name = "apple"
}
resource "garage_permission" "test" {
	access_key_id = garage_access_key.test.id
	bucket_id = garage_bucket.test.id
	owner = %t
	read = %t
	write = %t
}
`, owner, read, write)
}
