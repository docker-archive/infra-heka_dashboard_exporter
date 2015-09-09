# heka_exporter
heka_exporter is a very simple application that connects to the [heka dashboard](https://hekad.readthedocs.org/en/latest/config/outputs/dashboard.html) and exports the data
for prometheus consumption. Currently, the exporter is hard-coded to consume these heka sub-systems:
* decoders
* encoders
* globals
* outputs

You must enable the DashboardOutput in heka. Also, I recommend setting the interval to be something less
than your prometheus scraping interval.
```
[DashboardOutput]
ticker_interval = 15
```
Then, start the exporter like this:
```
./heka_exporter -heka.url="http://127.0.0.1:4352/data/heka_report.json"
```
Prometheus-compatible metrics will be available at http://127.0.0.1:9111/metrics

Or, to listen on an alternate port:
```
./heka_exporter -heka.url="http://127.0.0.1:4352/data/heka_report.json" -web.listen-address ":9999"
```
