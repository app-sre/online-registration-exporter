FROM registry.centos.org/centos/centos:7

COPY online-registration-exporter /bin/online-registration-exporter
COPY config.yml.sample            /etc/online-registration-exporter/config.yml

EXPOSE      9115
ENTRYPOINT  [ "/bin/online-registration-exporter" ]
CMD         [ "--config.file=/etc/online-registration-exporter/config.yml" ]
