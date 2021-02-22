python-multi-producers
======================

In this example we simulate multiple temperature sensors pushing their readings to Styx using the HTTP REST API. Styx easily scales to thousands of producers, though you'll probably have to raise your system's open file limit both for Styx and for the python script.

Usage
-----

Install requirements

```bash
pip3 install -r requirements.txt
```

Ensure Styx is running and create a "measures" log

```bash
curl localhost:8000/logs -X POST -d name=measures
```

Run 100 producers

```bash
python3 main.py 100
```

From another terminal, check that the log is filling with measures

```bash
curl localhost:8000/logs/measures
```
