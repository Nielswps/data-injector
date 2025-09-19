# Data Injector
A simplistic tool for injecting static data into different storage solutions, intended for local testing and showcasing.


# Usage
The tool can either be used as an executable or as a Docker image.

## Executing the tool directly
The tool can be executed with the `go run` command. The service needs an endpoint to post the data, so first ensure that a running instance of the target storage solution is running and accessible.
If you just want to test how the tool works, you can run the following command to inject some data into a Redit cache:

```shell
go run . --redis-address=127.0.0.1:6388 --data-file=./example/example_data.json
```

## Executing the tool through Docker
The tool can be build as a Docker image using the following command:

```shell
docker build . -t data-injector
```

## Data Format
The format of the data file is expected to be valid JSON and to follow the format below:
```json
[
    {
        "key": "key_1",
        "value": "val_1"
    },
    {
        "key": "key_2",
        "value": {"type": "object", "value": "some data"}
    },
    {
        "key": "key_3",
        "value": ["element_1", "element_2", "element_3"]
    }
]
```
The content of each "value" is parsed as a JSON value on its own and must therefore be valid JSON as well.