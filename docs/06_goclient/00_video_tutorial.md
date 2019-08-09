# Video Tutorial

## Installing Cadence Service and UI on Mac

{% include youtubePlayer.html id="aLyRyNe5Ls0" %}

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
