# Notes

## Compiling

Compilation is done by using the `ocb` binary in the repository (needs at least go 1.19).
```shell
./ocb --config builder-config.yaml
```

You can modify the `builder-config.yaml` to have as many custom processors, receivers and exporters as you need.

## Links

https://opentelemetry.io/docs/collector/trace-receiver/
https://opentelemetry.io/docs/collector/custom-collector/