---
layout: default
title: Getting started
permalink: /docs/tutorials/
---

# Getting started

## Installing Cadence Service and UI on a Mac

<figure class="video-container">
  <iframe
    src="https://www.youtube.com/embed/aLyRyNe5Ls0"
    frameborder="0"
    height="315"
    allowfullscreen
    width="560"></iframe>
</figure>

Commands executed during the tutorial:

```bash
docker-compose up

docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain samples-domain domain register

docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain samples-domain domain describe

alias cadence="docker run --rm ubercadence/cli:master --address host.docker.internal:7933"

cadence --domain samples-domain domain desc

cadence help

cadence workflow help

cadence --domain samples-domain workflow list

cadence --domain samples-domain workflow help start

cadence --domain samples-domain workflow start -wt test -tl test -et 300

cadence --domain samples-domain workflow list -op

cadence --domain samples-domain workflow terminate -wid <workflowID>

```
