package config

import (
	"bytes"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// Config represents the configuration for the exporter
type Config struct {
	Devices  []Device `yaml:"devices"`
	Features struct {
		BGP        bool `yaml:"bgp,omitempty"`
		DHCP       bool `yaml:"dhcp,omitempty"`
		DHCPL      bool `yaml:"dhcpl,omitempty"`
		DHCPv6     bool `yaml:"dhcpv6,omitempty"`
		Firmware   bool `yaml:"firmware,omitempty"`
		Health     bool `yaml:"health,omitempty"`
		Routes     bool `yaml:"routes,omitempty"`
		POE        bool `yaml:"poe,omitempty"`
		Pools      bool `yaml:"pools,omitempty"`
		Optics     bool `yaml:"optics,omitempty"`
		W60G       bool `yaml:"w60g,omitempty"`
		WlanSTA    bool `yaml:"wlansta,omitempty"`
		WlanIF     bool `yaml:"wlanif,omitempty"`
		Monitor    bool `yaml:"monitor,omitempty"`
		Ipsec      bool `yaml:"ipsec,omitempty"`
		QueueTrees bool `yaml:"queue_trees,omitempty"`
	} `yaml:"features,omitempty"`
}

// Device represents a target device
type Device struct {
	Name     string    `yaml:"name"`
	Address  string    `yaml:"address,omitempty"`
	Srv      SrvRecord `yaml:"srv,omitempty"`
	User     string    `yaml:"user"`
	Password string    `yaml:"password"`
	Port     string    `yaml:"port"`
}

type SrvRecord struct {
	Record string    `yaml:"record"`
	Dns    DnsServer `yaml:"dns,omitempty"`
}
type DnsServer struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// Load reads YAML from reader and unmashals in Config
func Load(r io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err = viper.ReadConfig(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	decodeHook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String {
			stringData := data.(string)
			if strings.HasPrefix(stringData, "${") && strings.HasSuffix(stringData, "}") {
				envVarValue := os.Getenv(strings.TrimPrefix(strings.TrimSuffix(stringData, "}"), "${"))
				if len(envVarValue) > 0 {
					return envVarValue, nil
				}
			}
		}
		return data, nil
	}

	c := &Config{}
	err = viper.Unmarshal(&c, viper.DecodeHook(decodeHook))
	if err != nil {
		return nil, err
	}

	return c, nil
}
