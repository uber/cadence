---
name: createsampledomain
---

Setup environment variables on Mac:
```bash
export CADENCE_CLI_DOMAIN=samples-domain
```

Setup environment variables on Windows:
```cmd
set USER=someusername
set CADENCE_CLI_DOMAIN=samples-domain
```

Check if samples-domain exists:
```bash
docker run --rm ubercadence/cli:master d desc
```

Create samples-domain:
```bash
docker run --rm ubercadence/cli:master d register --rd 7
```
