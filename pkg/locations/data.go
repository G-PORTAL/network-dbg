package locations

import (
	_ "embed"
	"github.com/charmbracelet/log"
	"strings"
)
import "github.com/go-yaml/yaml"

type LocationEndpoint string

func (l LocationEndpoint) Short() string {
	return strings.Split(string(l), ".")[0]
}

//go:embed locations.yml
var configData []byte

type config struct {
	Endpoints []LocationEndpoint `yaml:"location_endpoints"`
}

func GetLocations(limit *string) []LocationEndpoint {
	c := getConfig()

	locations := make([]LocationEndpoint, 0)
	for _, endpoint := range c.Endpoints {
		if limit != nil && *limit != "" && !strings.HasPrefix(string(endpoint), *limit) {
			continue
		}

		locations = append(locations, endpoint)
	}

	return locations
}

func getConfig() config {
	var cfg config
	err := yaml.Unmarshal(configData, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return cfg
}
