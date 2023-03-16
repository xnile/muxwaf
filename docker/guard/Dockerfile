FROM openresty/openresty:1.21.4.1-0-alpine-apk
COPY guard /opt/apps/muxwaf/guard

RUN chown -R nobody. /opt/apps/muxwaf \
    && ln -sf /dev/stdout /opt/apps/muxwaf/guard/logs/access.log \
    && ln -sf /dev/stdout /opt/apps/muxwaf/guard/logs/sampled.log \
    && ln -sf /dev/stderr /opt/apps/muxwaf/guard/logs/error.log

#CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
ENTRYPOINT ["/usr/local/openresty/bin/openresty", "-p", "/opt/apps/muxwaf/guard", "-c", "nginx.docker.conf", "-g", "daemon off;"]