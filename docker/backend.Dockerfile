FROM clamav/clamav:latest

HEALTHCHECK --interval=120s --start-period=120s --retries=5 CMD ["./clamdcheck.sh"]
