# microengine-clamav

This project provides a ClamAV-based microengine that can be used as a base to build more complicated microengines.
Users should be able to quickly get running by editing only one file with two functions to get started.

# Quickstart

## Install and test

Clone this repository.

`git clone https://github.com/polyswarm/microengine-clamav.git`

Install the package in development mode to allow customization. We recommend using a [virtual environment](https://docs.python.org/3/library/venv.html).

`pip install -e .[web,gunicorn,tests]`

**Note:** This engine requires a running ClamAV daemon (clamd). You can start one using Docker:

```bash
docker run -d -p 3310:3310 --name clamd clamav/clamav:latest
```

Wait for ClamAV to initialize (this may some time as it downloads virus definitions).

Check that the installation is working correctly

```console
$ python -m microengineclamav.engine analyze --check-eicar
{
  "metadata": {
    "product": "clamav",
    "scanner": {
      "version": "1.0",
      "environment": {
        "operating_system": "Linux",
        "architecture": "x86_64"
      }
    },
    "malware_family": "Eicar-Signature",
    "confidence": 1.0
  },
  "verdict": "malicious",
  "bid": 999900000000000000
}
```

Run the tests

```console
$ pytest -v
================== test session starts ==================
(...)
configfile: pyproject.toml
plugins: requests-mock-1.12.1, mock-3.15.1
collected 4 items

tests/test_scan.py::test_scan_file_malicious PASSED [ 25%]
tests/test_scan.py::test_scan_file_benign PASSED  [ 50%]
tests/test_server.py::test_valid_bounty_to_api PASSED [ 75%]
tests/test_server.py::test_invalid_bounty_to_api PASSED [100%]
============= 4 passed, 4 warnings in 0.09s =============
```

Now you have a working Engine that uses ClamAV to detect malware.

## Implementing your own engine

In `microengineclamav/engine.py` there some small functions.
The most important one is `analyze(bounty)`, which is where
to wire up the malware detection tool.

```py
# import polyswarm_engine as ps

@engine.register_analyzer
def analyze(bounty):
    with psengine.ArtifactTempfile(bounty) as path:
        stream_result = engine.cmd.scan_file(path)
        if len(stream_result) >= 2 and stream_result[0] == 'FOUND':
            result = {
                'verdict': psengine.MALICIOUS,
                'metadata': {'malware_family': stream_result[1].strip('\n')},
                'bid': psengine.bid_max(bounty),
            }
        else:
            result = {
                'verdict': psengine.BENIGN,
                'bid': psengine.bid_max(bounty),
            }
        return result


# Helper command, to split the logic and ease the debugging:
@engine.expose_command
def scan_file(filename):
    """Sends the `filename` to clamd for scanning"""
    with open(filename, 'rb') as file:
        # No need to close this.
        # Each connection is opened and closed on every call
        clamd_socket = _get_clamd_socket()
        result = clamd_socket.instream(file)
        return result.get('stream', []) if result else []
```

Your return dict will be checked against `polyswarm_engine.Analysis` rules,
e.g. a `verdict` is present and `metadata['confidence']` is a float
between 0.0 and 1.0 _if provided_.
For the full ruleset, have a peek at the `polyswarm_engine` codebase.

## Test your engine

During the implementation, you can issue ad-hoc tests calling
the `python -m microengineclamav.engine analyze` tool.
Alternatively, it also works by executing the file directly:

```console
$ cd microengineclamav
$ ./engine.py analyze --help
Usage: clamav analyze [OPTIONS] [ARTIFACTS]...

  Analyze artifacts

Options:
  -v, --verbose
  --check-empty     Verify this engine can analyze an empty
                                  bounty
  --check-eicar     Verify this engine can analyze EICAR test
                    file
  --check-wicar, --check-exploit-url
                    Verify this engine can analyze the WICAR
                    exploit kit URL
  -t, --artifact-type [bounty|file|url]
                    Artifact type to use when constructing
                    bounties. 'bounty' loads manually
                    constructed bounties, treating each argument
                    as the path to a JSON-encoded bounty object
--help              Show this message and exit.
```

The returned value will be checked for structure.

This CLI can issue scans for files in your disk, for local testing purposes:

```console
$ ./engine.py analyze ~/Downloads/Firefox\ Installer.exe
{
  ...
  "verdict": "benign",
  "bid": 999900000000000000
}
```

We recommend that you always check scans for:
- EMPTY bounties
- EICAR if creating a file-scanning engines
- WICAR if creating a url-scanning engine
- Return UNKNOWN for unsupported file types

## Example: Checking the file type

If you run an analysis for WICAR the template implementation will return BENIGN:

```console
$ ./engine.py analyze --check-wicar
{
  ...
  "verdict": "benign",
  "bid": 999900000000000000
}
...
AssertionError: Received 'benign' instead of malicious
```

As an example, for handling URL bounties gracefully,
you can change the `engine.py` file to have these new lines:

```diff
# import polyswarm_engine as ps

 @engine.register_analyzer
 def analyze(bounty):
+    if not psengine.bounty.is_file_artifact(bounty):
+        logger.error("Received non-file artifact bounty")
+        return psengine.bounty.UNSUPPORTED
     with psengine.ArtifactTempfile(bounty) as path:
         stream_result = engine.cmd.scan_file(path)
```

It will now change to answer non-file bounties with an UNSUPPORTED verdict.

```console
$ ./engine.py analyze --check-wicar
2025-10-22 20:30:19,022 - ERROR [engine.py:28][analyze] Received non-file artifact bounty
{
  ...
  "verdict": "unknown",
  "bid": 0
}
...
AssertionError: Received 'unknown' instead of malicious
```

Which is fine for a ClamAV engine, that is not supposed to handle URL bounties.

# Where to go from here?

This simple engine now does everything in the correct way.
Your existing malware-detection tool can be freely integrated within `engine.py`.

To help you get started, tooling exists inside the `polyswarm_engine` package.
For example, ClamAV can natively scan files on the filesystem via CLI.
`polyswarm_engine` have a context manager function that downloads the file
and stores in a temporary folder, making your life easier:

```py
# import polyswarm_engine as ps

@engine.register_analyzer
def analyze(bounty):
    with psengine.ArtifactTempfile(bounty) as path:
        # path is a temporary file path on the filesystem
        stream_result = engine.cmd.scan_file(path)
    ...
```

That and other niceties are covered in full on the [PolySwarm Documentation](https://docs.polyswarm.io/engines),
specially on the PolySwarm Engine Package section: https://docs.polyswarm.io/engines/polyswarm-engine-package/

# How it works?

During the tests above the `engine.py analyze` tool simulated a Bounty
already received and enqueued for processing inside a Celery worker.
Then it calls the `analyze()` function with that Bounty "dict".

For real engines, PolySwarm will send events as HTTP POST requests
to your server webhook, configured in the PolySwarm website.
Engines need to listen passively until a new event arrives.

Your webserver will receive HTTP requests. A python WSGI application running
handles the requests and enqueues a job to be processed by a worker.

The worker runs your function `analyze()` and it decides the appropriate response.
In the same job the worker sends the response back to PolySwarm.

# How to run this for real?

More details about the workflow briefly explained above, recommendations and
alternatives for common scenarios are also available
in the [PolySwarm Documentation](https://docs.polyswarm.io/engines).
