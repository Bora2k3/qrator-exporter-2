## Prometheus Exporter for QRATOR Labs

### Target service

  * [QRATOR Labs](https://client.qrator.net);

### Features

Fetches some metrics from Qrator API methods such as:
  *  [statistics_current_ip](https://api.qrator.net/#domain-methods-statistics-statistics-current-ip);
  *  [statistics_current_http](https://api.qrator.net/#domain-methods-statistics-statistics-current-http);
  
### Existing command line arguments and environment variables

Arguments

| Long name | Short name | Description | Default  |
| --------- | ---------- | ----------- | -------- |
| --qrator.client-id | -c | Your personal dashboard ID which obtained in dashboard. | 1 |
| --web.listen-address | -l | Address to listen on for web interface and telemetry. | :9805 |
| --web.telemetry-path | -p | Path to expose metrics. | /metrics |

Environment variable

  * QRATOR_TOKEN_AUTH;
  
NOTE! Create user account in Qrator Labs project and [make new API key](https://client.qrator.net/qrator/apitoken/) for access to metrics.

### How to run exporter

It's just example, take correct client ID from your user account
```bash
$ ./qrator-metrics --qrator.client-id cl0000
```

### How to run docker image as container

```bash
$ docker build --tag qrator-metrics .
$ docker run -itd --rm -p9805:9805 -e QRATOR_TOKEN_AUTH=<token> qrator-metrics -c cl0000
```
