FROM scratch
ADD conf/ /app/conf/
ADD bin/qb_101_lhd /app/bin/
WORKDIR /app
EXPOSE 8001
EXPOSE 8002
ENTRYPOINT ["/app/bin/qb_101_lhd"]