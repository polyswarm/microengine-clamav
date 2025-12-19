#!/usr/bin/env python3
__version__ = '1.0'

import logging
import os
import platform
import time

import clamd
import polyswarm_engine as psengine

logger = logging.getLogger(__name__)

engine = psengine.EngineManager(
    name='clamav',
    vendor='clamav',
    config={
        'CLAMD_HOST': os.getenv('CLAMD_HOST', 'localhost'),
        'CLAMD_PORT': int(os.getenv('CLAMD_PORT', '3310')),
        'CLAMD_TIMEOUT': int(os.getenv('CLAMD_TIMEOUT', '30')),
    },
)


def _get_clamd_socket() -> clamd.ClamdNetworkSocket:
    host_or_sock = engine.config['CLAMD_HOST']
    timeout = engine.config.get('CLAMD_TIMEOUT', '') or None
    if host_or_sock.startswith(('/', './', '../')):
        # Is a file socket.
        clamd_socket = clamd.ClamdUnixSocket(path=host_or_sock, timeout=timeout)
    else:
        # Is a network location
        clamd_socket = clamd.ClamdNetworkSocket(
            host_or_sock,
            engine.config['CLAMD_PORT'],
            timeout,
        )
    return clamd_socket


@engine.expose_command
def version():
    """Fetches the clamd server version"""
    sock = _get_clamd_socket()
    logger.info('Clamd socket: %r', sock)
    return sock.version()


@engine.expose_command
def scan_file(filename):
    """Sends the `filename` to clamd for scanning"""
    with open(filename, 'rb') as file:
        # No need to close this. Each connection is opened and closed in each method
        clamd_socket = _get_clamd_socket()
        result = clamd_socket.instream(file)
        return result.get('stream', []) if result else []


@engine.expose_command
def wait_clamd(tries=6):
    logger.warning('Waiting for clamd to be available at %s', engine.config['CLAMD_HOST'])
    for i in range(1, tries + 1):
        try:
            _get_clamd_socket().ping()
            break
        except clamd.ConnectionError as err:
            logger.warning('No connection to %s: %r', engine.config['CLAMD_HOST'], err)
            time.sleep(engine.config['CLAMD_TIMEOUT'])
            logger.debug('retrying: number %s', i)
    else:
        logger.error('Clamd socket on %s never got ready. Giving up.', engine.config['CLAMD_HOST'])


@engine.register_head
def head():
    vendor = engine.cmd.version()
    return {
        'scanner': {
            'version': __version__,
            'vendor_version': vendor.strip('\n'),
            'environment': {
                'operating_system': platform.system(),
                'architecture': platform.machine(),
            },
        }
    }


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


if __name__ == '__main__':
    engine.cli()
