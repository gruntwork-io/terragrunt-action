# Docker image to run terragrunt

Docker image with TGEnv and TGSwitch installed inside, which can be used to install and run Terragrunt.

Example usage:
```
tfenv install "1.4.6"
tfenv use "1.4.6"
TG_VERSION="0.46.3" tgswitch

terragrunt ...
```

## References

* https://github.com/tfutils/tfenv
* https://github.com/warrensbox/tgswitch