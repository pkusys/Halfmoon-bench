FROM shengqipku/halfmoon:sosp-ae as halfmoon
FROM golang:1.14-buster as builder

COPY --from=halfmoon /src/halfmoon /halfmoon
COPY . /workflow/workloads

RUN cd /workflow/workloads/beldi && \
    make hotel-baseline && \
    make media-baseline && \
    make retwis-baseline && \
    make singleop-baseline
RUN cd /workflow/workloads/boki && \
    make hotel && \
    make media && \
    make retwis && \
    make rw && \
    make singleop && \
    make recovery
RUN cd /workflow/workloads/optimal && \
    make hotel && \
    make media && \
    make retwis && \
    make rw && \
    make singleop && \
    make switching && \
    make recovery

FROM ubuntu:focal

COPY --from=builder  /workflow/workloads/beldi/bin  /beldi-bin
COPY --from=builder  /workflow/workloads/boki/bin   /bokiflow-bin
COPY --from=builder  /workflow/workloads/optimal/bin /optimal-bin
COPY --from=halfmoon     /boki/launcher       /boki/launcher
