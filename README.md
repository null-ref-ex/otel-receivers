# Notes

## Compiling

Compilation is done by using the `ocb` binary in the repository (needs at least go 1.19).
```shell
./ocb --config builder-config.yaml
```

You can modify the `builder-config.yaml` to have as many custom processors, receivers and exporters as you need.

## Local Dev

If you want to include changes to the receiver for local testing you need to push to a branch at remote and take the new commit hash and put it into the `builder-config.yaml` as the version of the receiver to pull.

## Links

https://opentelemetry.io/docs/collector/trace-receiver/
https://opentelemetry.io/docs/collector/custom-collector/