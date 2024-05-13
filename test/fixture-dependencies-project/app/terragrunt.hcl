dependency "module1" {
  config_path = "../module1"
  mock_outputs = {
    file = "mock1.txt"
  }
}

dependency "module2" {
  config_path = "../module2"
  mock_outputs = {
    file = "mock2.txt"
  }
}

inputs = {
  str1  = "46521694"
  num2  = 42
  file1 = dependency.module1.outputs.file
  file2 = dependency.module1.outputs.file
  repo_root = get_path_to_repo_root()
}
