# JSONSpec

The jsonspec package offers a simple way to define a spec (or schema) for JSON data. We use it to
validate the arguments for connectors.


## Usage

The usual way to use it is to define a struct type for the expected data. For example:

    type DatabaseSource struct {
        AuthMethod string `description:"Authentication method" enum:"password,iam" config:"credential" required:"true"`
        User       string `config:"credential" required:"true"`
        Password   string `config:"credential" tags:"secret" visible_when:"auth_method=password"`
        Host       string `config:"connection" required:"true"`
        Port       int    `config:"connection" default:"5432" advanced:"true"`
        Schedule   string `config:"schedule" default:"0 * * * *"`
    }

For this type the library will expect a JSON object like the following:

    {
        "auth_method": "password",
        "user": "alice",
        "password": "hunter2",
        "host": "db.example.com",
        "port": 5433,
        "schedule": "0 */6 * * *"
    }

The library will automatically convert between the different conventions for field names, for
example turning `AuthMethod` into `auth_method`.

You can generate a spec for this type as follows:

    spec, err := jsonspec.For(new(DatabaseSource))

Now you can call `spec.ValidateJSON` to check if a JSON document matches the spec. It'll return an
error if any of the fields have the wrong type or if any fields marked as required are missing.

You can also call `LoadJSON` to load a DatabaseSource from JSON:

    var source DatabaseSource
    err := jsonspec.LoadJSON(data, &source) // data is a []byte

This will generate a spec for DatabaseSource, validate that the input matches the spec, and store
the data in `source`.


## Struct tags

The following tags are recognized on struct fields.

Tags that affect input validation:

- `required:"true"`: the field must be provided in the JSON as a non-zero value.
- `enum:"a,b,c"`: list of allowed values. Only enforced for string fields.

Tags that only annotate the generated spec for downstream tooling (such as the connector
management frontend) and are ignored by `Validate`:

- `description:"..."`: human-readable description of the field.
- `default:"..."`: default value for the field (parsed according to its type).
- `tags:"a,b,c"`: list of free-form tags attached to the field.
- `config:"credential|connection|schedule"`: configuration layer the field belongs to. Empty
  means the field is not user-facing.
- `advanced:"true"`: field should be hidden behind an "advanced" toggle in the frontend for
  connector management.
- `visible_when:"<field>=<value>"`: field only renders in the frontend for connector
  management when the referenced field matches the given value.


## Spec as JSON

The Spec type is written so it can be marshaled and unmarshaled with `encoding/json`. Here's what
the spec for the DatabaseSource type above would look like:

    {
        "type": "object",
        "fields": {
            "auth_method": {
                "type": "string",
                "description": "Authentication method",
                "enum": ["password", "iam"],
                "required": true,
                "config": "credential"
            },
            "user": {
                "type": "string",
                "required": true,
                "config": "credential"
            },
            "password": {
                "type": "string",
                "tags": ["secret"],
                "config": "credential",
                "visible_when": "auth_method=password"
            },
            "host": {
                "type": "string",
                "required": true,
                "config": "connection"
            },
            "port": {
                "type": "integer",
                "default": 5432,
                "config": "connection",
                "advanced": true
            },
            "schedule": {
                "type": "string",
                "default": "0 * * * *",
                "config": "schedule"
            }
        }
    }
