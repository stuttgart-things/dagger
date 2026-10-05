# Reads a file that is NOT in this directory: the smoke test hands it in with
# --extra-files, the way a provider gets "${path.module}/ca.crt".
output "extra" {
  value = trimspace(file("${path.module}/extra.txt"))
}
