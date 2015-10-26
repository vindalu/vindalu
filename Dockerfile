
FROM centos:7

ADD ./build/vindaloo*.x86_64.rpm /root/

RUN cd /root && yum -y install *.rpm && ln -s /opt/vindaloo/bin/vindaloo /usr/local/bin/
RUN rm -rvf /root/vindaloo*.x86_64.rpm

EXPOSE 5454 4223 8223 6223

ENTRYPOINT ["vindaloo"]
