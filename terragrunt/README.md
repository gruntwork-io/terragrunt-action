# Docker image to run terragrunt

Docker image with [`mise`](https://mise.jdx.dev/) installed inside, which can be used to install and run Terragrunt.

Example usage:
```
mise use terraform@1.4.6
mise use opentofu@1.6.2
mise use terragrunt@0.46.3

terragrunt ...
```

## References

* https://mise.jdx.dev/


## Multiarch build notes

```
docker buildx create --name multiarch-builder
docker buildx use multiarch-builder

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t gruntwork/terragrunt:0.2.0-test-arm \
  --push \
  .

```
