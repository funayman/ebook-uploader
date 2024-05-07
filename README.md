# Simple eBook Uploader

Created as a way to lazily upload eBooks on my NAS instead of using `curl` every time.

Workflow:
1. Upload using Calibre Web button
2. POST data to uploader service
3. Save data to /uploads directory
4. /uploads directory watched by Calibre
5. Book added to library

## Example Docker Compose

```yaml
services:
  uploader:
    build: ./uploader
    environment:
      - UPLOAD_DIR=/uploads
      - UPLOAD_MAX_FILE_SIZE=1GB
    ports:
      - 8000:8000
    restart: unless-stopped
    depends_on:
      - calibre-web
    volumes:
      - ./calibre/upload:/uploads
    networks:
      - ebooks
      - web

  calibre-web:
    image: lscr.io/linuxserver/calibre-web:latest
    container_name: calibre-web
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Detroit
      - DOCKER_MODS=linuxserver/mods:universal-calibre #optional
      - OAUTHLIB_RELAX_TOKEN_SCOPE=0 #optional
    volumes:
      - ./web/config:/config
      - $HOME/Documents/CalibreLibrary:/books
    ports:
      - 8083:8083
    restart: unless-stopped
    depends_on:
      - calibre
    networks:
      - ebooks
      - web

  calibre:
    image: lscr.io/linuxserver/calibre:latest
    container_name: calibre
    security_opt:
      - seccomp:unconfined #optional
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Etc/UTC
      - PASSWORD=supersecret
      - USERNAME=drt
      - CLI_ARGS= #optional
    volumes:
      - ./calibre/config:/config
      - ./calibre/upload:/uploads
      - ./calibre/plugins:/plugins
      - $HOME/Documents/CalibreLibrary:/books
    ports:
      - 8080:8080 # vnc http
      # - 8181:8181 # vnc https
      - 8081:8081 # calibre web server
      - 9090:9090 # wireless device connection
    restart: unless-stopped
    networks:
      - ebooks
      - web

networks:
  ebooks:
    external: false
  web:
    external: true
```

## Example Caddyfile for Reverse Proxy

```
books.nas {
  handle /upload {
    reverse_proxy uploader:8000
  }

  vars PORT "8083"
  reverse_proxy calibre-web:{vars.PORT}
  tls {
    dns cloudflare {env.CLOUDFLARE_API_TOKEN}
  }
}
```

## Code Structure
```
.
├── cmd
│   ├── logfmt
│   │   └── main.go
│   └── server
│       ├── build
│       │   └── all
│       │       └── all.go
│       ├── handler
│       │   └── uploadgrp
│       │       ├── routes.go
│       │       └── uploadgrp.go
│       └── main.go
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── upload
│   ├── stores
│   │   ├── uploadfs
│   │   │   └── uploadfs.go
│   │   ├── uploadgcs
│   │   ├── uploadmulti
│   │   └── uploads3
│   └── upload.go
└── web
    ├── context.go
    ├── cors.go
    ├── debug
    │   └── debug.go
    ├── mid
    │   ├── basicauth.go
    │   ├── errors.go
    │   ├── limitbody.go
    │   ├── logger.go
    │   ├── mid.go
    │   └── panics.go
    ├── middleware.go
    ├── mux
    │   └── mux.go
    ├── request.go
    ├── response.go
    ├── signal.go
    └── web.go

18 directories, 27 files
```
