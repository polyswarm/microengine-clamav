import clamd
import io
import platform

from microengineclamav.models import Bounty, ScanResult, Verdict
from microengineclamav import settings

from polyswarmartifact.schema import ScanMetadata
from polyswarmartifact.artifact_type import ArtifactType


SYSTEM = platform.system()
MACHINE = platform.machine()


def scan(bounty: Bounty) -> ScanResult:
    content = bounty.fetch_artifact()
    # No need to close this. Each connection is opened and closed in each method
    clamd_socket = clamd.ClamdNetworkSocket(settings.CLAMD_HOST, settings.CLAMD_PORT, settings.CLAMD_TIMEOUT)
    try:
        vendor = clamd_socket.version()
        result = clamd_socket.instream(io.BytesIO(content))
    except clamd.ConnectionError as err:
        logger.exception('Error connecting to clamd')
        raise errors.CalledProcessScanError('Unable to connect') from err
    except clamd.ResponseError as err:
        logger.exception('Error in clamd')
        raise errors.CalledProcessScanError('Bad response') from err

    stream_result = result.get('stream', [])
    metadata = ScanMetadata().set_malware_family('')\
                             .set_scanner(operating_system=SYSTEM,
                                          architecture=MACHINE,
                                          vendor_version=vendor.strip('\n'))
    if len(stream_result) >= 2 and stream_result[0] == 'FOUND':
        metadata.set_malware_family(stream_result[1].strip('\n'))
        return ScanResult(verdict=Verdict.MALICIOUS, confidence=1.0, metadata=metadata)

    return ScanResult(verdict=Verdict.BENIGN, metadata=metadata)


def compute_bid(bounty: Bounty, scan_result: ScanResult) -> int:
    max_bid = bounty.rules.get(settings.MAX_BID_RULE_NAME, settings.DEFAULT_MAX_BID)
    min_bid = bounty.rules.get(settings.MIN_BID_RULE_NAME, settings.DEFAULT_MIN_BID)

    bid = min_bid + max(scan_result.confidence * (max_bid - min_bid), 0)
    bid = min(bid, max_bid)
    return bid
