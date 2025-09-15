# Data Injector
A simplistic tool for injecting static data into different storage solutions, intended for local testing and showcasing.


## Usage

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