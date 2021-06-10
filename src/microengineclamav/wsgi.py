from flask import Flask
from logging.config import dictConfig
from microengineclamav import settings

from microengineclamav.middleware import ValidateSenderMiddleware
from microengineclamav.views import api

dictConfig(settings.LOGGING)

app = Flask(__name__)
app.register_blueprint(api, url_prefix='/')
application = ValidateSenderMiddleware(app)
