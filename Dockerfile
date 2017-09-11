FROM neale/eris

RUN apk --no-cache add lua5.2
RUN apk --no-cache add lua5.3

# Install MOTH. This could be less obtuse.
COPY www /moth/www
COPY bin /moth/bin
RUN mkdir -p /moth/state/teams /moth/state/points.new /moth/state/points.tmp
RUN chown www:www /moth/state/teams /moth/state/points.new /moth/state/points.tmp
RUN mkdir -p /moth/packages
RUN touch /moth/state/points.log
RUN ln -s ../state/puzzles.json /moth/www/puzzles.json
RUN ln -s ../state/points.json /moth/www/points.json

COPY src/moth-init /usr/sbin/moth-init

WORKDIR /
CMD ["/usr/sbin/moth-init"]

