# -*- coding: utf-8 -*-
from setuptools import setup, find_packages

try:
    long_description = open("README.rst").read()
except IOError:
    long_description = ""

# -*- coding: utf-8 -*-
from setuptools import setup, find_packages

try:
    long_description = open("README.rst").read()
except IOError:
    long_description = ""

setup(
    name="microengineclamav",
    version="0.2.0",
    description="Clamav Engine",
    long_description=long_description,
    long_description_content_type="text/markdown",
    license="MIT",
    author="PolySwarm Developers",
    author_email="info@polyswarm.io",
    url='https://github.com/polyswarm/microengine-clamav',
    install_requires=[
        "celery",
        "clamd",
        "Flask",
        "microengine-utils",
        "polyswarm-artifact",
        "requests",
        "python-json-logger"
    ],
    packages=find_packages('src'),
    package_dir={'': 'src/'},
    python_requires='>=3.6.5,<4',
    extras_require={
        "tests": [
                'pytest',
                'pytest-cov',
                'pytest-mock',
                'requests-mock'
        ],
    },
    classifiers=[
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
    ]
)
