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
  x = "1"
  z = "xxxx"

  path = get_repo_root()
}
