resource "proxmoxve_acme_account" "test" {
  name      = "acme_test"
  contact   = "foo@bar.com"
  directory = "https://127.0.0.1:14000/dir"
  tos_url   = "foobar"
}
