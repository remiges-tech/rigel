package rigelctl

const RigelSchemaJSON = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "fields": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "type": {
            "type": "string",
            "enum": ["int", "float", "string", "bool"]
          },
          "description": {
            "type": "string"
          },
          "constraints": {
            "type": "object",
            "properties": {
              "min": {
                "type": "number"
              },
              "max": {
                "type": "number"
              },
              "enum": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              }
            },
            "additionalProperties": false
          }
        },
        "required": ["name", "type", "description"],
        "additionalProperties": false
      }
    },
    "description": {
      "type": "string"
    }
  },
  "required": ["fields", "description"],
  "additionalProperties": false
}`
