Media types
-----------

Three media types are available to deal with log records over HTTP within styx.

### Binary record

`application/octet-stream` 

This allows to specify that an HTTP body (request or response) is processed as one record.

### Binary records

`application/vnd.styx.binary-records`

To allow multiples records in an HTTP body a simple binary format can be used.

```
  +----------------+--------------------------------+
  |  size (int32)  |      record (size bytes)       |
  +----------------+--------------------------------+
```

Each record must be prefixed by a size, a big-endian int32, encoding the record length. 

### Line delimited records

`application/vnd.styx.line-delimited;line-ending=lf`

This media type provides an handy format when dealing with text records delimited with line endings, such as JSON entries for example.  
An optionnal media type param `line-ending` allows to specify expected line ending among following values `lf`, `cr` or `crlf`.  
The default is `lf`.

Note that the final line ending is mandatory.