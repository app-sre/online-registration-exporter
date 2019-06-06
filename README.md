# online-registration-exporter

Prometheus exporter for OpenShift Online Registration app

## Running this software

### From binaries

Download the most suitable binary from [the releases tab](https://github.com/app-sre/online-registration-exporter/releases)

Then:

    ./online-registration-exporter <flags>

### Using the container image

    docker run --rm -d -p 9115:9115 --name online-registration-exporter -v $PWD/config.yml:/etc/online-registration-exporter/config.yml quay.io/app-sre/online-registration-exporter:latest --config.file=/etc/online-registration-exporter/config.yml

## Building the software

### Local Build

    make build

### Building docker image

    make image image-push

## Configuration

online-registration-exporter is configured via a config.yml file and command-line flags (such as what configuration file to load, what port to listen on, and the logging format and level).

online-registration-exporter can reload its configuration file at runtime. If the new configuration is not well-formed, the changes will not be applied.
A configuration reload is triggered by sending a `SIGHUP` to the online-registration-exporter process or by sending a HTTP POST request to the `/-/reload` endpoint.

To view all available command-line flags, run `./online-registration-exporter -h`.

To specify which configuration file to load, use the `--config.file` flag.

Additionally, an [example configuration](config.yml.sample) is also available.

## License

Apache License 2.0, see [LICENSE](LICENSE).

