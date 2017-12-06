[![Build Status](https://travis-ci.org/pokidovea/mimicro.svg?branch=master)](https://travis-ci.org/pokidovea/mimicro) [![codecov](https://codecov.io/gh/pokidovea/mimicro/branch/master/graph/badge.svg)](https://codecov.io/gh/pokidovea/mimicro)

# mimicro

A simply configurable mock-server which allows you to prepare environment for integration testing of your application. You only need to create a config file with set of required servers and responses of their endpoints.

## Installation

You can download a prepared package [here](https://dl.equinox.io/pokidovea/mimicro/stable) or clone the repository and compile by yourself.

## Update

Once you've installed (or compiled) the application, you can check for updates by calling

```shell
mimicro -update
```

To see current version type 

```shell
mimicro -version
```

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

## Check config

```shell
mimicro -config config.yaml -check
```

## Run

```shell
mimicro -config config.yaml
```

After that you can make requests on `localhost:4573/simple/url` and get `{"some": "value"}` responses.

## Management server

The management server can be accessed on port `4444` by default. You can change this port by passinng a flag `-management-port <your port>`. The management server provides you with some useful tools, such as statistics of requests.

## Statistics of requests

After passing a flag `-collect-statistics` you can get statistics of the requests by address `localhost:4444/statistics/get?server=<server name>&url=<url like in the config>&method=<method in any case>`. All parameters are optional.

In order to reset statistics make a GET request to `localhost:4444/statistics/reset?server=<server name>&url=<url like in the config>&method=<method in any case>`. All parameters are optional too.

