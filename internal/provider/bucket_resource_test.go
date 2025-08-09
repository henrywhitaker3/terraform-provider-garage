package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccBucketResource(t *testing.T) {
	garage, cancel := garage(t)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: garage + testAccBucketResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_bucket.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("bongo"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "garage_bucket.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: garage + testAccBucketResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"garage_bucket.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("bongo"),
					),
				},
			},
		},
	})
}

func testAccBucketResourceConfig() string {
	return `
resource "garage_bucket" "test" {
	name = "bongo"
}
`
}
