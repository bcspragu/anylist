# (Br)AnyList

(Br)AnyList is a mobile-friendly (mobile-only?) frontend for viewing and
modifying lists on AnyList.

The backend code is heavily inspired by https://github.com/codetheweb/anylist

## Screenshot

It looks like this:

![Screenshot of a basic grocery list](/assets/screenshot.png)

## The Extremely Mundane Story

I received an AnyList subscription as a Christmas gift, but I'm not a big fan
of downloading apps. The AnyList website is nice, but it doesn't work well on
mobile. My needs are simple, a Google Keep-like add/view/remove items from a
list type of functionality. That's what this project will attempt to provide.

## What Works

The basics: Authentication, loading list data, adding/removing/checking/
unchecking items. Checked items go to the end.

## What Doesn't Work

Anything else. Specifically:

- [ ] Any sort of live updating
- [ ] Any sort of reordering
- [ ] Categories
- [ ] Prices
- [ ] Photos

With the exception of live updating, I probably won't even attempt to add any
other functionality.

## Deployment

### Building Release Artifacts

You can build the frontend and the backend as Docker containers.

For the backend:

```bash
# From the root directory
docker build -t your.docker.host/anylist .
```

For the frontend:
```bash
cd frontend
pnpm build:cloud
docker build -t your.docker.host/anylist-fe .
```

### Running

Personally, I run the two containers behind an NGINX server (also in Docker)
using basic auth, using something like this:

For running the containers:
```bash
# Create a network for the frontend to reach the backend
docker network create some-net

# Run the backend. I bake an age-encrypted sops [1] file into my Docker image
# and then mount the decryption key.
# [1] https://github.com/mozilla/sops#encrypting-using-age
docker run --network=some-net \
  --restart=always \
  --name anylist \
  -e "SOPS_AGE_KEY_FILE=/secrets/keys.txt" \
  -v /var/lib/srv/anylist/keys.txt:/secrets/keys.txt:ro \
  <my Docker registry>/anylist

# Run the frontend, point it at the backend's address on the Docker network.
docker run --network=some-net \
  --restart=always \
  --name anylist-fe \
  -e "PUBLIC_API_BASE_URL=http://anylist:8080" \
  <my Docker registry>/anylist-fe
```
  
Here's an example NGINX config, should go in a .conf file somewhere.

```nginx
server {
  listen 80;
  listen [::]:80;

  # Redirect HTTP -> HTTPS
  server_name <domain>;
  return 301 https://<domain>$request_uri;
}

server {
  listen 443 ssl;
  listen [::]:443 ssl;
  server_name <domain>;

  # TLS config with LetsEncrypt
  include /etc/nginx/letsencrypt.conf;
  ssl_certificate /etc/letsencrypt/live/<domain cert>/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/<domain cert>/privkey.pem;

  satisfy any;

  # Don't allow any traffic except basic auth.
  deny all;

  # Basic auth user/pass live in an htpasswd file
  auth_basic "forbidden";
  auth_basic_user_file /etc/nginx/conf.d/htpasswd;

  # Default traffic to the Svelte app
  location / {
    proxy_set_header  X-Forwarded-For   $remote_addr;
    proxy_set_header  X-Forwarded-Proto $scheme;

    resolver 127.0.0.11 valid=10s;
    set $upstreamName anylist-fe:3000;
    proxy_pass                          http://$upstreamName;
    proxy_set_header  Host              $http_host;
    proxy_set_header  X-Real-IP         $remote_addr; # pass on real client's IP
  }

  # Forward API requests onto the backend directly,
  location /api/ {
    proxy_set_header  X-Forwarded-For   $remote_addr;
    proxy_set_header  X-Forwarded-Proto $scheme;

    resolver 127.0.0.11 valid=10s;
    set $upstreamName anylist:8080;
    proxy_pass                          http://$upstreamName;
    proxy_set_header  Host              $http_host;
    proxy_set_header  X-Real-IP         $remote_addr; # pass on real client's IP
  }
}
```
