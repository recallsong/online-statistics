FROM alpine:3.6

RUN mkdir -p /online-statistics/conf && mkdir -p /online-statistics/page
Copy bin/linux-amd64-online-statistics /online-statistics/online-statistics
Copy conf/config-example.yml /online-statistics/conf/config.yml
Copy page /online-statistics/page

EXPOSE 80 8610
WORKDIR /online-statistics
VOLUME /online-statistics/conf

CMD ["/online-statistics/online-statistics"]