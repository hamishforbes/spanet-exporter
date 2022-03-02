package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/hamishforbes/spanet_exporter/spanet_client"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Namespace = "spanet"
)

type Exporter struct {
	username     string
	passwordHash string
	spaName      string
	logger       log.Logger
	client       *spanet_client.SpaConn

	up                *prometheus.Desc
	water_temp        *prometheus.Desc
	target_temp       *prometheus.Desc
	heating           *prometheus.Desc
	auto              *prometheus.Desc
	sanitising        *prometheus.Desc
	cleaning          *prometheus.Desc
	sleeping          *prometheus.Desc
	blower_mode       *prometheus.Desc
	blower_speed      *prometheus.Desc
	lights_active     *prometheus.Desc
	lights_mode       *prometheus.Desc
	lights_brightness *prometheus.Desc
	lights_speed      *prometheus.Desc
	lights_colour     *prometheus.Desc
	pump_active       *prometheus.Desc
	pump_ok           *prometheus.Desc
	locked            *prometheus.Desc
}

func New(username string, passwordHash string, spaName string, logger log.Logger) *Exporter {
	client := spanet_client.New(username, passwordHash)
	if err := client.Login(); err != nil {
		logger.Log("msg", "Error logging in", "err", err)
		return nil
	}

	conn, err := client.Connect(spaName)
	if err != nil {
		logger.Log("msg", "Error Connecting to spa", "err", err)
	}

	return &Exporter{
		username:     username,
		passwordHash: passwordHash,
		spaName:      spaName,
		logger:       logger,
		client:       &conn,
		up: prometheus.NewDesc(
			"up",
			"Could the spa be reached.",
			[]string{"spa_name"},
			nil,
		), water_temp: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "water_temp"),
			"Current water temperature.",
			[]string{"spa_name"},
			nil,
		), target_temp: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "target_temp"),
			"Target water temperature.",
			[]string{"spa_name"},
			nil,
		), heating: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "heating"),
			"Heater is on.",
			[]string{"spa_name"},
			nil,
		),
		auto: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "auto"),
			"Auto is enabled.",
			[]string{"spa_name"},
			nil,
		),
		sanitising: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "sanitising"),
			"Sanitise cycle is active.",
			[]string{"spa_name"},
			nil,
		),
		cleaning: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "cleaning"),
			"UV / Ozone is active.",
			[]string{"spa_name"},
			nil,
		),
		sleeping: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "sleeping"),
			"Sleep mode is active.",
			[]string{"spa_name"},
			nil,
		),
		blower_mode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "blower_mode"),
			"Blower Mode. 2 when off, 1 when in ramp mode and 0 in variable mode.",
			[]string{"spa_name"},
			nil,
		),
		blower_speed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "blower_speed"),
			"Blower Speed. 1 to 5.",
			[]string{"spa_name"},
			nil,
		),
		lights_active: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "lights_active"),
			"Lights on or off.",
			[]string{"spa_name"},
			nil,
		),
		lights_mode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "lights_mode"),
			"Lighting mode. 0 for white, 1 for colour, 2 for step, 3 for fade, 4 for party mode!",
			[]string{"spa_name"},
			nil,
		),
		lights_brightness: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "lights_brightness"),
			"Lights brightness. 1 to 5.",
			[]string{"spa_name"},
			nil,
		),
		lights_speed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "lights_speed"),
			"Light effect speed. 1 to 5.",
			[]string{"spa_name"},
			nil,
		),
		lights_colour: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "lights_colour"),
			"Light colour. 0 to 30.",
			[]string{"spa_name"},
			nil,
		),
		pump_ok: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "pump_ok"),
			"Pump is OK to turn on.",
			[]string{"spa_name", "pump"},
			nil,
		),
		pump_active: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "pump_active"),
			"Pump is on.",
			[]string{"spa_name", "pump"},
			nil,
		),
		locked: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "locked"),
			"Control panel is locked. 0 for unlocked, 1 for partial, 2 for full lock.",
			[]string{"spa_name"},
			nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.water_temp
	ch <- e.target_temp
	ch <- e.heating
	ch <- e.auto
	ch <- e.sanitising
	ch <- e.cleaning
	ch <- e.sleeping
	ch <- e.blower_mode
	ch <- e.blower_speed
	ch <- e.lights_active
	ch <- e.lights_mode
	ch <- e.lights_brightness
	ch <- e.lights_speed
	ch <- e.lights_colour
	ch <- e.pump_active
	ch <- e.pump_ok
	ch <- e.locked
}

// Implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	up := float64(1)

	if e.client.Conn == nil {
		e.logger.Log("msg", "Reconnecting to Spa")
		err := e.client.Connect()
		if err != nil {
			e.logger.Log("msg", "Error connecting to spa", "err", err)
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, e.spaName)
		}
	}

	spa, err := e.client.Read()
	if err != nil {
		e.logger.Log("msg", "Error reading data from spa", "err", err)
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0, e.spaName)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, up, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.water_temp, prometheus.GaugeValue, spa.WaterTemperature, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.target_temp, prometheus.GaugeValue, spa.TargetTemperature, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.heating, prometheus.GaugeValue, spa.Heating, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.auto, prometheus.GaugeValue, spa.Auto, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.sanitising, prometheus.GaugeValue, spa.Sanitising, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.cleaning, prometheus.GaugeValue, spa.Cleaning, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.sleeping, prometheus.GaugeValue, spa.Sleeping, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.blower_mode, prometheus.GaugeValue, spa.Blower.Mode, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.blower_speed, prometheus.GaugeValue, spa.Blower.Speed, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.lights_active, prometheus.GaugeValue, spa.Lights.Active, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.lights_mode, prometheus.GaugeValue, spa.Lights.Mode, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.lights_brightness, prometheus.GaugeValue, spa.Lights.Brightness, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.lights_speed, prometheus.GaugeValue, spa.Lights.Speed, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.lights_colour, prometheus.GaugeValue, spa.Lights.Colour, e.spaName)
	ch <- prometheus.MustNewConstMetric(e.locked, prometheus.GaugeValue, spa.Settings.Lock, e.spaName)

	for _, pump := range spa.Pumps {
		ch <- prometheus.MustNewConstMetric(e.pump_active, prometheus.GaugeValue, pump.Active, e.spaName, fmt.Sprint(pump.Id))
		ch <- prometheus.MustNewConstMetric(e.pump_ok, prometheus.GaugeValue, pump.Ok, e.spaName, fmt.Sprint(pump.Id))
	}

}
