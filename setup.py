# -*- coding: utf-8 -*-
from setuptools import setup, find_packages

try:
    long_description = open("README.rst").read()
except IOError:
    long_description = ""

setup(
    name="microengineclamav",
    version="0.1.0",
    description="Clamav microengine",
    license="MIT",
    author="PolySwarm Developers",
    packages=find_packages('src'),
    package_dir={'': 'src/'},
    install_requires=[
        "celery",
        "Flask",
        "requests",
        "python-json-logger"
    ],
    long_description=long_description,
    classifiers=[
        "Programming Language :: Python",
        "Programming Language :: Python :: 3.8",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ]
)