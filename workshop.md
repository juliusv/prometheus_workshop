# Getting started

## Getting Prometheus
Download the latest binary release of Prometheus for your platform from:

https://github.com/prometheus/prometheus/releases

Extract the contents into a new directory and change to that directory.

Example for Linux:

```
mkdir prometheus
cd prometheus
wget https://github.com/prometheus/prometheus/releases/download/0.13.2/prometheus-0.13.2.linux-amd64.tar.gz
tar xfvz prometheus-0.13.2.linux-amd64.tar.gz
```

## Configuring Prometheus to monitor itself
Download the example configuration:

```
wget https://raw.githubusercontent.com/prometheus/prometheus/master/documentation/examples/prometheus.yml
```

Take a look at the configuration file. It configures global options, as well as
a single job to scrape metrics from: the Prometheus server itself.

Prometheus collects metrics from monitored targets by scraping metrics HTTP
endpoints on these targets. Since Prometheus also exposes data in the same
manner about itself, it may also be used to scrape and monitor its own health.
While a Prometheus server which collects only data about itself is not very
useful in practice, it is a good starting example.

## Starting Prometheus
Start Prometheus. By default, Prometheus stores its database in /tmp/metrics
(flag `-storage.local.path`).

```
./prometheus -config.file=prometheus.conf
```

Prometheus should start up and it should show a status page about itself at
[http://localhost:9090/](http://localhost:9090/). In the *Status* section, you
will find [http://localhost:9090/metrics](http://localhost:9090/metrics) in the
list of scraped targets. Give Prometheus a couple of seconds to start
collecting data about itself from its own HTTP metrics endpoint.

You can also verify that Prometheus is serving metrics about itself by
navigating to its metrics exposure endpoint:
[http://localhost:9090/metrics](http://localhost:9090/metrics).

By default, Prometheus will only use at most one OS thread at a time. In
production scenarios on multi-CPU machines, you will most likely achieve better
performance by setting the `GOMAXPROCS` environment variable to a value similar
to the number of available CPU cores:

```
GOMAXPROCS=8 ./prometheus -config.file=prometheus.conf
```

## Using the expression browser
The query interface at
[http://localhost:9090/graph](http://localhost:9090/graph) allows you to
explore metric data collected by the Prometheus server. At the moment, the
server is only scraping itself. The collected metrics are already quite
interesting, though.  The *Console* tab shows the most recent value of metrics,
while the *Graph* tab plots values over time. The latter can be quite expensive
(for both the server and the browser). It is in general a good idea to try
potentially expensive expressions in the *Console* tab first. Take a bit of
time to play with the expression browser. Suggestions:

* Evaluate `prometheus_local_storage_ingested_samples_total`, which shows you
  the total number of ingested samples over the lifetime of the server. In the
  *Graph* tab, it will show as steadily increasing.
* `The expression prometheus_local_storage_ingested_samples_total[1m]`
  evaluates to all sample values of the metric in the last minute. It cannot be
  plotted as a graph, but in the *Console* tab, you see a list of the values with
  (Unix) timestamp.
* `rate(prometheus_local_storage_ingested_samples_total[1m])` calculates the
  rate (increase per second) over the 1m timeframe. In other words, it tells you
  how many samples per second your server is ingesting. This expression can be
  plotted nicely, and it will become more interesting as you add more targets.

## Start the node exporter
The node exporter is a server which exposes system statistics about the machine
it is running on as Prometheus metrics.

Download the latest node exporter binary release for your platform from:

https://github.com/prometheus/node_exporter/releases

Beware that the majority of the node exporter's functionality is
Linux-specific, so its exposed metrics will be significantly reduced when
running it on other platforms.

Linux example:

```
mkdir node_exporter
cd node_exporter
wget https://github.com/prometheus/node_exporter/releases/download/0.8.1/node_exporter-0.8.1.linux-amd64.tar.gz
tar xvfz node_exporter-0.8.1.linux-amd64.tar.gz
```

Start the node exporter:

```
./node_exporter
```

## Configure Prometheus to monitor the node exporter
Add the following job configuration to your prometheus.conf to monitor both your own and the provided node exporter:

```
job: {
  name: "node"
  scrape_interval: "15s"
  target_group: {
    target: "http://localhost:9100/metrics"
  }
}
```

If you are not running your local node exporter under Linux, you might want to
point your Prometheus server to a Linux node exporter run by one of your peers
in the workshop. Or point it to a node exporter we are running during the
workshop at http://demo-node.prometheus.io:9100/metrics.

Stop your Prometheus server (e.g. kill it with Ctrl+C) and start it up again.
Config changes still require a server restart to take effect, but this will no
longer be the case starting at Prometheus version 0.14.0.

Then check the *Status* page of your Prometheus server to make sure the node
exporter is scraped correctly. Shortly after, a whole lot of interesting
metrics will show up in the expression browser, each of them starting with
`node_`. (Reload the page to see them in the autocompletion.) As an example,
have a look at `node_cpu`.

The node exporter has a whole lot of modules to export machine metrics. Have a
look at https://github.com/prometheus/node_exporter to get an idea. While
Prometheus is particularly good at collecting service metrics, correlating
those with system metrics from individual machines can be immensely helpful.
(Perhaps that one task that showed high latency yesterday was scheduled on a
node with a lot of competing disk operations?)

## Use the node exporter to export the contents of a text file
The *textfile* module of the node exporter can be used to expose static
machine-level metrics (such as what role a machine has) or the outcome of
machine-tied batch jobs (such as a Chef client run). To use it, create a
directory for the text files to export and (re-)start the node exporter with
the `--collector.textfile.directory` flag set. Finally, create a text file in
that directory.

```
mkdir textfile-exports
./node_exporter --collector.textfile.directory=textfile-exports
echo 'role{role="srecon_node_exporter"} 1' > textfile-exports/role.prom.$$
mv textfile-exports/role.prom.$$ textfile-exports/role.prom
```

For details, see https://github.com/prometheus/node_exporter#textfile-collector.

## Configuring multiple targets per job
In a target group, multiple targets can be configured. For example, you might
want to add a node exporter of a peer in the workshop:

```
job: {
  name: "node"
  scrape_interval: "15s"
  target_group: {
    target: "http://localhost:9100/metrics"
    target: "http://192.168.1.113:9100/metrics"
  }
}
```

Prometheus adds an `instance` label with the hostname and port as the value to
each metric scraped from any target. With that label, you can later aggregate
or separate metrics from different targets.

In practice, configuring many targets statically is often a maintenance burden.
The solution is service discovery. Currently, Prometheus supports service
discovery via DNS SRV records, but support for more methods of service
discovery is planned for version 0.14.  To try out a DNS SRV record, we have
created one for `demo-node.prometheus.io`:

```
dig +short SRV demo-node.prometheus.io
```

Only one host and port is returned (the already known `demo-node.prometheus.io`
on port 9100), but any number of host/port combinations could be part of the
SRV record. Prometheus regularly polls the DNS information and dynamically
adjusts the targets. To configure a job with DNS service discovery, add the
following to `prometheus.conf`:

```
job {
  name: "node"
  sd_name: "demo-node.prometheus.io" # The SRV name.
  metrics_path: "/metrics" # Need it separate as is not part of the SRV record.
}
```

# The expression language

With more metrics collected by your Prometheus server, it is time to
familiarize yourself a bit more with the expression language. For comprehensive
documentation, see http://prometheus.io/docs/querying/basics/. The following is
meant as an inspiration for how to play with the metrics currently collected by
your server. Evaluate them in the *Console* and *Graph* tab. For the latter, try
different time ranges and the *stacked* option.

## The `rate()` function
Prometheus internally organizes sample data in chunks. It performs a number of
different chunk operations on them and exposes them as
`prometheus_local_storage_chunk_ops_total`, which is comprised of a number of
counters, one per possible chunk operation. To see a rate of chunk operations
per second, use the rate function over a time range that should cover at least
a handful of scrape intervals.

```
rate(prometheus_local_storage_chunk_ops_total[1m])
```

Now you can see the rate for each chunk operation type.  Note that the rate
function handles counter resets (for example if a binary is restarted).
Whenever a counter goes down, the function assumes that a counter reset has
happened and the counter has started counting from `0`.

## The `sum` aggregation operator
If you want to get the total rate for all operations, you need to sum up the
rates:

```
sum(rate(prometheus_local_storage_chunk_ops_total[1m]))
```

Note that you need to take the sum of the rate, and not the rate of the sum.
(Exercise for the reader: Why?)

## Select by label
If you want to look only at the persist operation, you can filter by label with
curly braces:

```
rate(prometheus_local_storage_chunk_ops_total{type="persist"}[1m])
```

You can use multiple label pairs within the curly braces (comma-separated), and
the match can be inverted (with `!=`) or performed with a regular expression
(with `=~`, or `!~` for the inverted match).

(Exercise: How to estimate the average number of samples per chunk?)

## Aggregate by label
The metric `http_request_duration_microseconds_count` counts the number of HTTP
requests processed. (Disregard the `duration_microseconds` part for now. It
will be explained later.) If you look at it in the *Console* tab, you can see
the many time series with that name. The metric is partitioned by handler,
instance, and job, resulting in many sample values at any given time. We call
that an instant vector.

If you are only interested in which job is serving how many QPS, you can let
the sum operator aggregate by job (resulting in the two jobs we are monitoring,
the Prometheus itself and the node exporter):

```
sum(rate(http_request_duration_microseconds_count[5m])) by (job)
```

A combination of label pairs is possible, too. You can aggregate by job and
instance (which is interesting if you have added an additional node exporter to
your config):

```
sum(rate(http_request_duration_microseconds_count[5m])) by (job, instance)
```

Note that there is an alternative syntax with the `by` clause following
directly the aggregation operator. This syntax is particularly useful in
complex nested expressions, where it otherwise becomes difficult to spot which
`by` clause belongs to which operator.

```
sum by (job, instance) (rate(http_request_duration_microseconds_count[5m]))
```

## Arithmetic
There is a metric `http_request_duration_microseconds_sum`, which sums up the
duration of all HTTP requests. If the labels match, you can easily divide two
instant vectors, yielding the average request duration in this case:

```
rate(http_request_duration_microseconds_sum[5m]) / rate(http_request_duration_microseconds_count[5m])
```

You can aggregate as above if you do it separately for numerator and
denominator:

```
sum(rate(http_request_duration_microseconds_sum[5m])) by (job) / sum(rate(http_request_duration_microseconds_count[5m])) by (job)
```

Things become more interesting if the labels do not match perfectly between two
instant vectors or you want to match vector elements in a many-to-one or
one-to-many fashion. See
http://prometheus.io/docs/querying/operators/#vector-matching for details.

## Summaries
Rather than an average request duration, you will be more often interested in
quantiles like the median or the 90th percentile. To serve that need,
Prometheus offers summaries. `http_request_duration_microseconds` is a summary
of HTTP request durations, and `http_request_duration_microseconds_sum` and
`http_request_duration_microseconds_count` are merely byproducts of that
summary.  If you look at `http_request_duration_microseconds` in the expression
browser, you see a multitude of time series, as the metric is now partitioned
by quantile, too. An expression like
`http_request_duration_microseconds{quantile="0.9"}` displays the 90th
percentile request duration. You might be tempted to aggregate the result as
you have done above. Not possible, unfortunately. Welcome to the wonderland of
statistics.

Read more at http://prometheus.io/docs/practices/histograms/.

## Recording rules
In your practical work with Prometheus at scale, you will pretty soon run into
expressions that are very expensive and slow to evaluate. The remedy is
*recording* rules, a way to tell Prometheus to pre-calculate expressions,
saving the result in a new time series, which can then be used instead of the
expensive expression. See
http://prometheus.io/docs/querying/rules/#recording-rules for details and
http://prometheus.io/docs/practices/rules/ for crucial best practices around
naming rules.

# Instrument code: Go
*This section is about instrumenting a Go application. If you prefer Python, continue with the next section.*

## The example application
Get the example application from GitHub:

```
$ cd $GOPATH/src/
$ mkdir -p github.com/juliusv
$ cd github.com/juliusv
$ git clone https://github.com/juliusv/prometheus_workshop.git
$ cd prometheus_workshop
$ go get
$ go build
$ ./prometheus_workshop
```

Study the code to understand what it is doing. Note that the application is for
demonstration purposes and implements a server and a client in the same binary.

## Instrument it
Instrument the server part with Prometheus. Things to keep in mind:

* What would be useful to instrument?
* What would be good variable names?
* How can I instrument in one place rather than many?
* How can/should I use labels?
* How to expose the /metrics endpoint?

The Prometheus Go client library is documented at https://godoc.org/github.com/prometheus/client_golang/prometheus. You will find valuable guidelines about Prometheus instrumentation in general at http://prometheus.io/docs/practices/instrumentation/. Also have a look at the naming guidelines http://prometheus.io/docs/practices/naming/.

If you are lost, you can look at instrumented code in the branch called
`instrumented` in the GitHub repository above. Note that the example
instrumentation is not necessarily ideal and/or complete.

# Instrument Code: Python
*This section is about instrumenting a Python application. If you prefer Go,
continue with the previous section.*

## The example application
Get the example application from GitHub:

```
$ pip install prometheus_client
$ git clone https://github.com/brian-brazil/prometheus_workshop_python
```

If you don't want to install python libraries globally, pass the `--user` flag to pip.

## Instrument it
Instrument the client and server with Prometheus. Things to keep in mind:

* What would be useful to instrument?
* What would be good variable names?
* How can I instrument in one place rather than many?
* How can/should I use labels?
* How to expose the /metrics endpoint?

The Prometheus Python client library is documented at
https://github.com/prometheus/client_python#prometheus-python-client. You will
find valuable guidelines about Prometheus instrumentation in general at
http://prometheus.io/docs/practices/instrumentation/. Also have a look at the
naming guidelines http://prometheus.io/docs/practices/naming/.

# Dashboard Building: Console Templates
Console templates are a built-in dashboarding system in the Prometheus server.
They are based on Go's templating language, which is more strongly typed than a
typical web templating engine.

You can see an example at
[http://localhost:9090/consoles/node.html](http://localhost:9090/consoles/node.html).

Task: Create a dashboard of QPS, latency, and "up" servers for the Go/Python
code you instrumented above.

Use
https://github.com/prometheus/prometheus/blob/master/consoles/cassandra.html as
a base to work off.

# Dashboard Building: PromDash
PromDash is a browser-based dashboard builder for Prometheus. It is a Rails
application and stores its dashboard metadata in a configurable SQL backend.
The actual graph data is retrieved by the browser via AJAX requests from the
configured Prometheus servers.

For the purposes of this workshop, we provide a shared PromDash server for you
at http://demo-node.prometheus.io:3000/. It already has one Prometheus server
configured as a data source: [http://localhost:9090/](http://localhost:9090/).
This server should work for everyone to build dashboards against their locally
running Prometheus servers (if you browse the web from a different host than
where you are running your test Prometheus server, you might need to create a
new server entry at http://demo-node.prometheus.io:3000/servers).

Let's create a dashboard to monitor the health of the Prometheus instance
itself:

1. Head over to http://demo-node.prometheus.io/3000 and click "New Dashboard".
2. Create a dashboard called "<username>-workshop" (you don't need to select a
   directory). PromDash will redirect you to your new, empty dashboard.
3. Set the "Range" input field just under the dashboard title to "30m" to show
   the last 30 minutes of data in the dashboard (feel free to play with the graph
   time range later).

Let's create a graph that shows the ingested samples per second:

1. Click on the "Datasources" menu item in the header line of the empty graph.
2. Click "Add Expression" and set the expression to
   `rate(prometheus_local_storage_ingested_samples_total[1m])`
   The graph should show the per-second rate of ingested samples.
3. Let's give the graph a title. Open the "Graph and axis settings" graph menu
   and set the title to "Ingested samples [rate-1m]".
4. Open the "Legend Settings" graph menu and set "Show legend" to "never",
   since this graph only contains a single time series.
5. Press "Save Changes" to save your progress.

Let's add another graph showing the rates of the various chunk operations:

1. Click the "Add Graph" button to add a second graph.
2. Add the following expression to the graph:

   `rate(prometheus_local_storage_chunk_ops_total[1m])`

   The graph should now show the per-second rate of chunk operations of various kinds.
3. Set the graph title to "Chunk ops [rate-1m]".
4. The legend currently shows all labels of the returned time series, although
   only the "chunk" label differs. To show only that label in the legend, click
   the "Legend Settings" tab and set the existing "Legend format" input to
   `{{type}}`.
5. Because a graph may have multiple expressions with different applicable
   legend format strings each, we still need to assign each legend format string
   to a particular expression. Open the "Datasources" graph menu again and in the
   "- Select format string -" dropdown, select the format string that you just
   created.
6. Press "Save Changes" to save your progress.

Finally, let's add a gauge that shows the number of expression queries
performed against your Prometheus server per second:

1. Click the "Add Gauge" button to add a gauge.
2. Set the gauge expression to:

   `scalar(sum(rate(http_request_duration_microseconds_count{handler=~"/api/query"}[1m])))`

   The gauge should now show the number of expression queries per second. Note
   that a gauge only supports queries which result in scalar values without any
   labels. Thus, the entire expression is wrapped in a scalar() call.
3. Under the "Chart settings" menu tab of your gauge, set the title to
   "Expression queries [rate-1m]".
4. Let's adjust the gauge's maximum value. In the "Chart settings" menu tab of
   the gauge, set the "Gauge Max" to a lower value that seems reasonable for the
   rate of queries your server is getting. For example, try setting it to 5.
5. Press "Save Changes" to save your progress.

Your dashboard should now look somewhat like this:

[![PromDash screenshot](/images/promdash.png)](#promdash)

PromDash supports many more features which we will not be able to explore in
this workshop. For example:

* graphing multiple expressions from different servers in the same graph
* mapping expressions to different axes and setting various axis options
* building templatized dashboards using template variables
* adding pie charts, Graphite graphs, or arbitrary web content
* etc.

For a more comprehensive overview, see the [PromDash
documentation](http://prometheus.io/docs/visualization/promdash/).

# Alerting
With instrumentation and a meaningful dashboard in place, the time is ripe to
think about alerting.  Alerting rules are set up similarly to recording rules.
See documentation at http://prometheus.io/docs/querying/rules/#alerting-rules.
You can inspect the status of configured alerts in the Alerts section of the
Prometheus server's status page
[http://localhost:9090/alerts](http://localhost:9090/alerts). However, for
proper notifications, you need to set up an Alertmanager. See
https://github.com/prometheus/alertmanager.

To play with the Alertmanager, you can download a release from
https://github.com/prometheus/alertmanager/releases, build one from source
yourself, or use the demo Alertmanager we have set up for the duration of the
workshop at http://demo-node.prometheus.io:9093. Note that with multiple
workshop participants sending alerts to the same Alertmanager, you might run
into naming collisions.  To point a Prometheus server to an Alertmanager, use
the `-alertmanager.url` flag.

For meaningful alerting, refer to the best practices section about alerting:
http://prometheus.io/docs/practices/alerting/.

Create a useful alerting rule for your example application, and then change the
code to simulate an outage triggering the alert.

# Pushing Metrics
Occasionally, you might need to push metrics that are not machine-related. (The
latter would be exposed via the `textfile` module of the node exporter, see
above.) The Pushgateway is a possible solution in that case. Check out the
details at http://prometheus.io/docs/instrumenting/pushing/. Note that this
does not change Prometheus's semantics to a push-based model.

To play with the Pushgateway, you can download a release from
https://github.com/prometheus/pushgateway/releases, build one from source
yourself, or use the demo Pushgateway we have set up for the duration of the
workshop at http://demo-node.prometheus.io:9091. Note that with multiple
workshop participants pushing to the Pushgateway, you will override each
other's metrics if you are using the same label values for job and instance.
