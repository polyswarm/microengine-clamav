import logging

from celery import Celery

from microengineclamav.models import Bounty, ScanResult, Verdict, Assertion, Phase
from microengineclamav import settings
from microengineclamav.scan import scan, compute_bid

celery_app = Celery('tasks', broker=settings.BROKER)
logger = logging.getLogger(__name__)


@celery_app.task
def handle_bounty(bounty):
    bounty = Bounty(**bounty)
    scan_result = scan(bounty)
    logger.debug('Bounty %s got ScanResult %s', bounty, scan_result)

    if bounty.phase == Phase.ARBITRATION:
        scan_response = scan_result.to_vote()
    else:
        if scan_result.verdict in [Verdict.UNKNOWN, Verdict.SUSPICIOUS]:
            # These results don't bid any NCT.
            bid = 0
        else:
            bid = compute_bid(bounty, scan_result)
        scan_response = scan_result.to_assertion(bid)

    bounty.post_response(scan_response)
