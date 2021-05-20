from celery import Celery

from microengineclamav.models import Bounty, ScanResult, Verdict, Assertion
from microengineclamav import settings
from microengineclamav.scan import scan, compute_bid

celery_app = Celery('tasks', broker=settings.BROKER)


@celery_app.task
def handle_bounty(bounty):
    bounty = Bounty(**bounty)
    scan_result = scan(bounty)
    bid = None

    if scan_result.verdict == Verdict.MALICIOUS or scan_result.verdict == Verdict.BENIGN:
        bid = compute_bid(bounty, scan_result)

    bounty.post_assertion(scan_result.to_assertion(bid))
