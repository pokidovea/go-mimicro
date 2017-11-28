[![Build Status](https://travis-ci.org/pokidovea/mimicro.svg?branch=master)](https://travis-ci.org/pokidovea/mimicro) [![codecov](https://codecov.io/gh/pokidovea/mimicro/branch/master/graph/badge.svg)](https://codecov.io/gh/pokidovea/mimicro)

# mimicro

A simply configurable mock-server which allows you to prepare environment for integration testing of your application. You only need to create a config file with set of required servers and responses of their endpoints.

## Config example

```yaml
# config.yaml

servers:
  - name: server_1
    port: 4573
    endpoints:
      - url: /simple/url
        GET:
          template: "{\"some\": \"value\"}"
          headers:
            content-type: application/json
        POST:
          template: "OK"
          status_code: 201
          headers:
            content-type: text/plain
      - url: /picture
        GET:
          file: file://mimicro.png
      - url: /{var}/in/filepath
        GET:
          file: file://{{.var}}micro.png
      - url: /template_from_file/{var}
        PUT:
          template: file://response_with_var.json
          headers:
            content-type: application/json
      - url: /string_template/{var}
        DELETE:
          template: "var is {{.var}}"
          status_code: 403
```

## How to run

```
> mimicro -config config.yaml -collect-statistics
```

After that you can make requests on `localhost:4573/simple/url` and get `{"some": "value"}` responses.
