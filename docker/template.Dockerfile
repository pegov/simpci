FROM docker:28.3-cli

RUN touch /tmp/entrypoint.sh

CMD ["/tmp/entrypoint.sh"]
