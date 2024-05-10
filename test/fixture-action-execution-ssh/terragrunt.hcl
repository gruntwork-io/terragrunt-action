inputs = {
  name = "World"
}

terraform {
  source = "git@github.com:gruntwork-io/terragrunt.git//test/fixture-download/hello-world?ref=v0.9.9"
}
