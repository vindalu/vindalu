
FROM centos:7

ADD ./build/vindalu*.x86_64.rpm /root/

RUN cd /root && yum -y install *.rpm && ln -s /opt/vindalu/bin/vindalu /usr/local/bin/
RUN rm -rvf /root/vindalu*.x86_64.rpm

EXPOSE 5454 4223 8223 6223

ENTRYPOINT ["vindalu"]
