# JSONSpec

The jsonspec package offers a simple way to define a spec (or schema) for JSON data. We use it to
validate the arguments for connectors.


## Usage

The usual way to use it is to define a struct type for the expected data. For example:

    type Person struct {
        ID        int    `description:"Unique ID" required:"true"`
        FirstName string
        LastName  string `required:"true"`
        Password  string `required:"true" tags:"secret"`
        IsAdmin   bool
    }

For this type the library will expect a JSON object like the following:

    {
        "id": 123,
        "first_name": "Jane",
        "last_name": "Doe",
        "password": "hunter2",
        "is_admin": true
    }

The library will automatically convert between the different conventions for field names, for
example tuning `FirstName` into `first_name`.

You can generate a spec for this type as follows:

    spec, err := jsonspec.For(new(Person))

Now you can call `spec.ValidateJSON` to check if a JSON document matches the spec. It'll return an
error if any of the fields have the wrong type or if any fields marked as required are missing.

You can also call `LoadJSON` to load a Person from JSON:

    var person Person
    err := jsonspec.LoadJSON(data, &person) // data is a []byte

This will generate a spec for Person, validate that the input matches the spec, and store the data
in `person`.


## Spec as JSON

The Spec type is written so it can be marshaled and unmarshaled with `encoding/json`. Here's what
the spec for the Person type would look like:

    {
        "type": "object",
        "fields": {
            "id": {
                "type": "int",
                "description": "Unique ID",
                "required": true
            },
            "first_name": {
                "type": "string"
            },
            "last_name": {
                "type": "string",
                "required": true
            },
            "password": {
                "type": "string",
                "required": true,
                "tags": [
                    "secret"
                ]
            },
            "is_admin": {
                "type": "boolean"
            }
        }
    }
