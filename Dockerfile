FROM python:3.12-slim

COPY ./run.py /root
COPY ./geth /root

WORKDIR /root

ENTRYPOINT python3 run.py
