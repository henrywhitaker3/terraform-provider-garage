resource "garage_bucket" "example" {
  name = "bongo"

}

resource "garage_access_key" "example" {
  name = "bongo"
}

resource "garage_permission" "example" {
  access_key_id = garage_access_key.example.id
  bucket_id     = garage_bucket.example.id
  read          = true
  write         = true
}
