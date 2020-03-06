FROM alpine:3.9

LABEL maintainer="Michael Senn <michael@morrolan.ch>"

COPY config-watcher /usr/local/bin/

# exec form did not work, throwing a `/bin/sh config-watcher not found` error
CMD 'config-watcher'
