package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/routeros.v2/proto"
	"strconv"
)

type queueTreeCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newQueueTreeCollector() routerOSCollector {
	c := &queueTreeCollector{}
	c.init()
	return c
}

func (c *queueTreeCollector) init() {
	c.props = []string{"name", "parent", "bytes", "packets", "dropped", "rate", "packet-rate", "queued-packets", "queued-bytes"}

	labelNames := []string{"name", "address", "queue", "comment"}
	c.descriptions = make(map[string]*prometheus.Desc)
	for _, p := range c.props[1:] {
		c.descriptions[p] = descriptionForPropertyName("queuetree", p, labelNames)
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

	for _, p := range c.props[2:] {
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
		ctx.ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, ctx.device.Name, ctx.device.Address, qt, comment)
	}
}
