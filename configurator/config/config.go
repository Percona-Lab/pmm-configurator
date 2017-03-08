package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

// PMMConfig implements struct with all configuration params in one place
type PMMConfig struct {
	ConfigPath         string              `yaml:"config"               default:""                        desc:"configuration file location"`
	HtpasswdPath       string              `yaml:"htpasswd-path"        default:"/srv/nginx/.htpasswd"    desc:"htpasswd file location"`
	ListenAddress      string              `yaml:"listen-address"       default:"127.0.0.1:7777"          desc:"Address and port to listen on: [ip_address]:port"`
	PathPrefix         string              `yaml:"url-prefix"           default:"/configurator"           desc:"Prefix for the internal routes of web endpoints"`
	SSHKeyPath         string              `yaml:"ssh-key-path"         default:""                        desc:"authorized_keys file location"`
	SSHKeyOwner        string              `yaml:"ssh-key-owner"        default:"admin"                   desc:"Owner of authorized_keys file"`
	GrafanaDBPath      string              `yaml:"grafana-db-path"      default:"/srv/grafana/grafana.db" desc:"grafana database location"`
	PrometheusConfPath string              `yaml:"prometheus-conf-path" default:"/etc/prometheus.yml"     desc:"prometheus configuration file location"`
	UpdateDirPath      string              `yaml:"update-dir-path"      default:"/srv/update"             desc:"update directory location"`
	Configuration      map[string]string   `yaml:"configuration"        default:""                        desc:""`
	Users              []map[string]string `yaml:"users"                default:""                        desc:""`
}

// ParseConfig implements function which read command line arguments, configuration file and set default values
func ParseConfig() (c PMMConfig) {
	t := reflect.TypeOf(&c).Elem()
	v := reflect.ValueOf(&c).Elem()
	// iterate over all confConfig fields
	for i := 0; i < v.NumField(); i++ {
		// get string with argument description, like "Owner of authorized_keys file"
		descTag := t.Field(i).Tag.Get("desc")
		// skip config file only fields
		if descTag != "" {
			// get poiner to confConfig field, like &c.SSHKeyOwner
			valueAddr := v.Field(i).Addr().Interface().(*string)
			// get string with argument name, like "ssh-key-owner"
			yamlTag := t.Field(i).Tag.Get("yaml")
			// pass pointer, argument name and argument description to flag library
			flag.StringVar(valueAddr, yamlTag, "", descTag)
		}
	}

	flag.Parse()
	c.parseConfig()
	flag.Parse() // command line should overide config
	c.setDefaultValues()

	return c
}

func (c *PMMConfig) setDefaultValues() {
	t := reflect.TypeOf(c).Elem()
	v := reflect.ValueOf(c).Elem()
	// iterate over all confConfig fields
	for i := 0; i < v.NumField(); i++ {
		// get string with current value of field (which have been read from config)
		curValue := v.Field(i).String()
		// get string with default value of field
		defValue := t.Field(i).Tag.Get("default")

		if curValue == "" {
			v.Field(i).SetString(defValue)
		}
	}
}

func (c *PMMConfig) parseConfig() {
	// parseConfig() runs before setDefaultValues(), so it is needed to set default manually
	if c.ConfigPath == "" {
		c.ConfigPath = "/srv/update/pmm-manage.yml"
	}

	configBytes, err := ioutil.ReadFile(c.ConfigPath)
	if os.IsNotExist(err) {
		// ignore config file is not exists
		return
	}
	if err != nil {
		log.Fatalf("Cannot read '%s' config file: %s\n", c.ConfigPath, err)
	}

	err = yaml.Unmarshal(configBytes, &c)
	if err != nil {
		log.Fatalf("Cannot parse '%s' config file: %s\n", c.ConfigPath, err)
	}
}

// Save dump configuration values to configuration file
func (c *PMMConfig) Save() {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("Cannot encode configuration: %s\n", err)
		return
	}

	if err = ioutil.WriteFile(c.ConfigPath, bytes, 0644); err != nil {
		log.Printf("Cannot save configuration file: %s\n", err)
		return
	}
}