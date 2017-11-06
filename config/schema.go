package config

var schema string = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "type": "object",
    "additionalProperties": false,
    "required": ["servers"],
    "properties": {
        "collectStats": {
            "type": "boolean"
        },
        "servers": {
            "uniqueItems": true,
            "items": {"$ref": "#/definitions/server"}
        }
    },
    "definitions": {
        "server": {
            "type": "object",
            "additionalProperties": false,
            "required": [
                "name",
                "port",
                "endpoints"
            ],
            "properties": {
                "name": {"type": "string"},
                "port": {"type": "integer"},
                "endpoints": {
                    "type": "array",
                    "uniqueItems": true,
                    "items": {"$ref": "#/definitions/endpoint"}
                }
            }
        },
        "endpoint": {
            "type": "object",
            "additionalProperties": false,
            "required": ["url"],
            "properties": {
                "url": {"type": "string"},
                "GET": {"$ref": "#/definitions/response"},
                "POST": {"$ref": "#/definitions/response"},
                "PUT": {"$ref": "#/definitions/response"},
                "PATCH": {"$ref": "#/definitions/response"},
                "DELETE": {"$ref": "#/definitions/response"}
            }
        },
        "response": {
            "type": "object",
            "additionalProperties": false,
            "required": ["body"],
            "properties": {
                "body": {"type": "string"},
                "content_type": {"type": "string"},
                "status_code": {
                    "type": "integer",
                    "enum": [
                        100, 101, 102,
                        200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
                        300, 301, 302, 303, 304, 305, 306, 307, 308,
                        400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410,
                        411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 424, 426, 428, 429, 431, 451,
                        500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511
                    ]
                }
            }
        }
    }
}
`
