resource "local_file" "f1" {
  content  = "file"
  filename = "${path.module}/file.txt"
}

output "file" {
  value = local_file.f1.filename
}