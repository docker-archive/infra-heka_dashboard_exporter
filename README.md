# heka_dashboard_exporter
heka_dashboard_exporter is a very simple application that connects to the [heka dashboard](https://hekad.readthedocs.org/en/latest/config/outputs/dashboard.html) and exports the data
for Prometheus consumption. Currently, the exporter is hard-coded to consume these heka sub-systems:
* decoders
* encoders
* filters
* globals
* outputs

You must enable the DashboardOutput in heka. Also, I recommend setting the interval to be something less
than your Prometheus scraping interval.
```
[DashboardOutput]
ticker_interval = 15
```
Then, start the exporter like this:
```
./heka_dashboard_exporter -heka.url="http://127.0.0.1:4352/data/heka_report.json"
```
Prometheus-compatible metrics will be available at http://127.0.0.1:9111/metrics

To listen on an alternate port, use the web.listen-address flag:
```
./heka_dashboard_exporter -heka.url="http://127.0.0.1:4352/data/heka_report.json" -web.listen-address=":9999"
```

# Docker container #

If you prefer, you can run the pre-built docker container with the same options as above:
```
docker run -d -p 9111:9111 dckr/heka_dashboard_exporter -heka.url="http://somehekainstance:4352/data/heka_report.json"
```
