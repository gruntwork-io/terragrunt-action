terraform {
  source = "git@github.com:gruntwork-io/terragrunt.git//test/fixture-download/hello-world?ref=v0.9.9"
}

locals {
  get_path_to_repo_root = get_path_to_repo_root()
  get_path_from_repo_root = get_path_from_repo_root()
}

inputs = {
  name = "Test git commands"
}
