FROM scratch

ADD pkg-linux-amd64 /pkg

ENTRYPOINT ["/pkg"]
