Getting started
---------------

### Quick Install

```bash
$ TODO git clone
$ cd styx
$ docker build -t styx .
```

### Run Styx

```bash
$ docker run -it --rm -p 8000:8000 --name styx styx
```

### Create a log

```bash
$ docker exec styx styx logs create myLog
name:                   myLog
status:                 ok
record_count:           0
file_size:              0
start_position:         0
end_position:           0
```

### Write records

```bash
$ docker exec -it styx styx logs write myLog
>my first record
>my second record
```

### Read records

```bash
$ docker exec styx styx logs read myLog
my first record
my second record
```
