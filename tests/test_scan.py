import base64
import dataclasses
import datetime

from microengineclamav.models import Bounty, Verdict, Assertion
from microengineclamav.utils import to_wei
from microengineclamav.tasks import handle_bounty

from polyswarmartifact.schema import Verdict as Metadata


EICAR_STRING = base64.b64decode(
    b'WDVPIVAlQEFQWzRcUFpYNTQoUF4pN0NDKTd9JEVJQ0FSLVNUQU5EQVJELUFOVElWSVJVUy1URVNULUZJTEUhJEgrSCo='
)


def test_scan_malicious(requests_mock, mocker):
    # Setup mock assertion
    spy = mocker.spy(Bounty, 'post_assertion')
    artifact_uri = 'mock://example.com/eicar'
    response_url = 'mock://example.com/response'
    eicar_sha256 = '275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f'
    # Setup http mocks
    requests_mock.get(artifact_uri, text=EICAR_STRING.decode('utf-8'))
    requests_mock.post(response_url, text='Success')

    bounty = Bounty(id=12345,
                    artifact_type='FILE',
                    artifact_uri=artifact_uri,
                    sha256=eicar_sha256,
                    mimetype='text/plain',
                    expiration=datetime.datetime.now().isoformat(),
                    phase='assertion_window',
                    response_url=response_url,
                    rules={
                        'max_allowed_bid': to_wei(1),
                        'min_allowed_bid': to_wei(1/16)
                    }
                    )

    handle_bounty(dataclasses.asdict(bounty))

    # Have to dissect the call, since we don't know what the metadata will be, it is version dependent.
    called_assertion = spy.mock_calls[0][1][1]
    assert called_assertion.verdict == Verdict.MALICIOUS.value


def test_scan_benign(requests_mock, mocker):
    # Setup mock assertion
    spy = mocker.spy(Bounty, 'post_assertion')
    artifact_uri = 'mock://example.com/not-eicar'
    response_url = 'mock://example.com/response'
    eicar_sha256 = '09688de240a0b492aca7af12057b7f24cd5d0439f14d40b9eec1ce920bc82cb6'
    # Setup http mocks
    requests_mock.get(artifact_uri, text='not-eicar')
    requests_mock.post(response_url, text='Success')

    bounty = Bounty(id=12345,
                    artifact_type='FILE',
                    artifact_uri=artifact_uri,
                    sha256=eicar_sha256,
                    mimetype='text/plain',
                    expiration=datetime.datetime.now().isoformat(),
                    phase='assertion_window',
                    response_url=response_url,
                    rules={
                        'max_allowed_bid': 1 * 10 ** 18,
                        'min_allowed_bid': 1 * 10 ** 18 / 16,
                    }
                    )

    handle_bounty(dataclasses.asdict(bounty))

    # Have to dissect the call, since we don't know what the metadata will be, it is version dependent.
    called_assertion = spy.mock_calls[0][1][1]
    assert called_assertion.verdict == Verdict.BENIGN.value
