package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strconv"
)

type queueTreeCollector struct {
	props []string

	descriptions map[string]*prometheus.Desc
	gauges       map[string]struct{}
}

func newQueueTreeCollector() routerOSCollector {
	c := &queueTreeCollector{}
	c.init()
	return c
}

func (c *queueTreeCollector) init() {
	c.props = []string{"bytes", "packets", "dropped", "rate", "packet-rate", "queued-packets", "queued-bytes"}
	helpText := []string{"Total Bytes", "Total Packets", "Dropped Packets", "Average Throughput Rate", "Average Packet Rate", "Queued Packets", "Queued Bytes"}

	labelNames := []string{"name", "address", "queue", "comment"}
	c.descriptions = make(map[string]*prometheus.Desc)
	for i, p := range c.props {
		c.descriptions[p] = descriptionForPropertyNameHelpText("queuetree", p, labelNames, helpText[i])
	}

	// Some data elements are gauges
	gaugeValues := []string{"rate", "packet-rate", "queued-bytes", "queued-packets"}
	c.gauges = make(map[string]struct{}, len(gaugeValues))
	for _, g := range gaugeValues {
		c.gauges[g] = struct{}{}
	}
}

func (c *queueTreeCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *queueTreeCollector) collect(ctx *collectorContext) error {
	stats, err := c.fetch(ctx)
	if err != nil {
		return err
	}

	for _, re := range stats {
		c.collectForStat(re, ctx)
	}

	return nil
}

func (c *queueTreeCollector) fetch(ctx *collectorContext) ([]*proto.Sentence, error) {
	reply, err := ctx.client.Run("/queue/tree/getall", "?disabled=false", "?invalid=false")
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching queue tree metrics")
		return nil, err
	}

	return reply.Re, nil
}

func (c *queueTreeCollector) collectForStat(re *proto.Sentence, ctx *collectorContext) {
	name := re.Map["name"]
	comment := re.Map["comment"]

	for _, p := range c.props {
		c.collectMetricForProperty(p, name, comment, re, ctx)
	}
}

func (c *queueTreeCollector) collectMetricForProperty(property, qt, comment string, re *proto.Sentence, ctx *collectorContext) {
	desc := c.descriptions[property]
	if value := re.Map[property]; value != "" {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"device":     ctx.device.Name,
				"queue_tree": qt,
				"property":   property,
				"value":      value,
				"error":      err,
			}).Error("error parsing queue tree metric value")
			return
		}

		_, isGauge := c.gauges[property]
		if isGauge {
			ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, ctx.device.Name, ctx.device.Address, qt, comment)
		} else {
			ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, ctx.device.Name, ctx.device.Address, qt, comment)
		}
	}
}
