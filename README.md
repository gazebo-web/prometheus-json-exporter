A Prometheus exporter that fetches arbitrary JSON from a remote API and exports the values as Prometheus gauge metrics.

## Installation

To use in other projects

```bash
go get gitlab.com/ignitionrobotics/web/prometheus-json-exporter
```

## Building

This project requires Go `1.10+` and `dep` for dependency management.

```bash
git clone https://gitlab.com/ignitionrobotics/web/prometheus-json-exporter
cd prometheus-json-exporter
dep ensure
go build
```

You can optionally install it

```bash
go install
```

### Image

A Dockerfile is provided to create a docker image. A built docker image is available in this repository's container
registry.

```bash
docker run registry.gitlab.com/ignitionrobotics/web/prometheus-json-exporter <target> [flags]
```

## Usage

```bash
prometheus-json-exporter <target> [flags]
```

### Parameters

| Parameter | Description | Example |
| --- | --- | --- | 
| `<target>` | Remote API URL to fetch JSON from. | `http://validate.jsontest.com/?json={"key":"value"}` |

### Flags

In addition, the following flags are available:

| Shorthand | Flag | Description | Default |
| --- | --- | --- | --- |
| -p | --prefix | Prefix added to all exported Prometheus metrics. Used to help make metrics unique. | `""` |
| -a | --listen-address | Address to listen in. This includes the host and port. | `:9116` |

### Endpoints

The `/metrics` endpoint will contact the API and return the JSON response as Prometheus metrics.

### Example

```bash
# Example JSON Data
curl -s "http://validate.jsontest.com/?json=%7B%22key%22:%22value%22%7D"
{
   "object_or_array": "object",
   "empty": false,
   "parse_time_nanoseconds": 24618,
   "validate": true,
   "size": 1
}

# Set the target and set an "example_" prefix to exported Prometheus metrics.
./prometheus-json-exporter http://validate.jsontest.com/?json=%7B%22key%22:%22value%22%7D --prefix example_

# Get the JSON as Prometheus metrics
curl -s "http://localhost:9116/metrics"
    # HELP example_empty Retrieved value
    # TYPE example_empty gauge
    example_empty 0
    # HELP example_parse_time_nanoseconds Retrieved value
    # TYPE example_parse_time_nanoseconds gauge
    example_parse_time_nanoseconds 41626
    # HELP example_size Retrieved value
    # TYPE example_size gauge
    example_size 1
    # HELP example_validate Retrieved value
    # TYPE example_validate gauge
    example_validate 1
```

## Attribution

This repository is a fork of 
[Shiroyagicorp's prometheus-json-exporter](https://github.com/shiroyagicorp/prometheus-json-exporter).
