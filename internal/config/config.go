package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultConfigPath = "config/config.yaml"
	defaultRoutesPath   = "config/routes.yaml"
)

// Root est le contenu de config/config.yaml.
type Root struct {
	Port       string `yaml:"port"`
	RoutesFile string `yaml:"routes_file"`
	Timeouts   struct {
		Default string `yaml:"default"`
	} `yaml:"timeouts"`
}

// routesYAML est le contenu de routes.yaml (liste brute avant validation).
type routesYAML struct {
	Routes []routeYAML `yaml:"routes"`
}

type routeYAML struct {
	Prefix   string `yaml:"prefix"`
	Upstream string `yaml:"upstream"`
	Timeout  string `yaml:"timeout,omitempty"`
}

// App regroupe la configuration résolue (prête pour le routeur / proxy).
type App struct {
	Port           string
	DefaultTimeout time.Duration
	Routes         []Route
}

// Route est une règle préfixe → service amont.
// Timeout à 0 : le proxy utilisera App.DefaultTimeout (si non nul).
type Route struct {
	Prefix   string
	Upstream *url.URL
	Timeout  time.Duration
}

// Load lit un fichier YAML de type Root (config gateway).
func Load(path string) (Root, error) {
	var out Root
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, fmt.Errorf("config: %w", err)
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("config yaml: %w", err)
	}
	return out, nil
}

// LoadApp charge config + routes depuis les chemins habituels ou les variables d'environnement.
// GATEWAY_CONFIG : fichier config racine (défaut config/config.yaml).
// ROUTES_FILE : surcharge du fichier routes (sinon routes_file du YAML, sinon config/routes.yaml).
func LoadApp() (*App, error) {
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}

	root, err := Load(cfgPath)
	if err != nil {
		return nil, err
	}

	routesPath := os.Getenv("ROUTES_FILE")
	if routesPath == "" {
		routesPath = root.RoutesFile
	}
	if routesPath == "" {
		routesPath = defaultRoutesPath
	}

	var defaultTimeout time.Duration
	if s := strings.TrimSpace(root.Timeouts.Default); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("timeouts.default: %w", err)
		}
		defaultTimeout = d
	}

	routes, err := loadRoutesFile(routesPath)
	if err != nil {
		return nil, err
	}

	port := strings.TrimSpace(root.Port)
	return &App{
		Port:           port,
		DefaultTimeout: defaultTimeout,
		Routes:         routes,
	}, nil
}

func loadRoutesFile(path string) ([]Route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("routes file: %w", err)
	}

	var doc routesYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("routes yaml: %w", err)
	}

	seen := make(map[string]struct{})
	out := make([]Route, 0, len(doc.Routes))
	for i, raw := range doc.Routes {
		prefix := strings.TrimSpace(raw.Prefix)
		if prefix == "" {
			return nil, fmt.Errorf("routes[%d]: prefix vide", i)
		}
		if prefix[0] != '/' {
			return nil, fmt.Errorf("routes[%d]: prefix %q doit commencer par /", i, prefix)
		}
		if _, ok := seen[prefix]; ok {
			return nil, fmt.Errorf("routes: prefix dupliqué %q", prefix)
		}
		seen[prefix] = struct{}{}

		upstreamStr := strings.TrimSpace(raw.Upstream)
		if upstreamStr == "" {
			return nil, fmt.Errorf("routes[%d]: upstream vide", i)
		}
		u, err := url.Parse(upstreamStr)
		if err != nil {
			return nil, fmt.Errorf("routes[%d].upstream: %w", i, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("routes[%d].upstream: URL invalide (schéma et hôte requis): %q", i, upstreamStr)
		}

		var rt time.Duration
		if s := strings.TrimSpace(raw.Timeout); s != "" {
			rt, err = time.ParseDuration(s)
			if err != nil {
				return nil, fmt.Errorf("routes[%d].timeout: %w", i, err)
			}
		}

		out = append(out, Route{
			Prefix:   prefix,
			Upstream: u,
			Timeout:  rt,
		})
	}

	return out, nil
}
