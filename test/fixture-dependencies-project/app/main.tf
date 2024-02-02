variable "str1" {}

variable "num2" {}

variable "file1" {}

variable "file2" {}

resource "local_file" "test_file" {
  content  = "${var.str1} ${var.num2} ${var.file1} ${var.file2}"
  filename = "${path.module}/test.txt"
}