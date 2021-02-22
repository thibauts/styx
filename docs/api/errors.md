Errors
------

Styx uses consistant error format accross its API. Errors are returned whith an HTTP error code from `4xx` to `5xx`.

Example

```json
  {
    "code": "log_exist",
    "message": "api: log already exists"
  }
```

`code` field contains an error code which can be used to react programmatically to a type of error.   
`message` field contains an human readable error message.