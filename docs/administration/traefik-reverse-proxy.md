# Use Traefik as a reverse proxy

You can set up Traefik in two ways:

<details>
<summary>Using Docker labels</summary>

Add these labels to your `docker-compose.yml` file:

```yml
labels:
  - traefik.http.routers.qingyugist.rule=Host(`qingyugist.example.com`) # Change to your subdomain
  # Uncomment the line below if you run QingyuGist in a subdirectory
  # - traefik.http.routers.app1.rule=PathPrefix(`/qingyugist{regex:$$|/.*}`) # Change opentist in the regex to yuor subdirectory name
  - traefik.http.routers.qingyugist.entrypoints=websecure # Change to the name of your 443 port entrypoint
  - traefik.http.routers.qingyugist.tls.certresolver=lets-encrypt # Change to certresolver's name
  - traefik.http.routers.qingyugist.service=qingyugist
  - traefik.http.services.qingyugist.loadBalancer.server.port=6157
```
</details>
<details>
<summary>Using a <code>yml</code> file</summary>

> [!Note]
> Don't forget to change the `<server-address>` to your server's IP

`traefik_dynamic.yml`
```yml
http:
  routers:
    qingyugist:
      entrypoints: websecure
      rule: Host(`qingyugist.example.com`) # Comment this line and uncomment the line below if using a subpath
      # rule: PathPrefix(`/qingyugist{regex:$$|/.*}`) # Change opentist in the regex to yuor subdirectory name
      # middlewares:
      #   - qingyugist-fail2ban
      service: qingyugist
      tls:
        certresolver: lets-encrypt
  services:
    qingyugist:
      loadbalancer:
        servers:
          - url: "http://<server-address>:6157"

```

</details>