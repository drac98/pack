package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/buildpacks/pack/internal/style"
)

type Config struct {
	RunImages           []RunImage       `toml:"run-images"`
	DefaultBuilder      string           `toml:"default-builder-image,omitempty"`
	DefaultRegistry     string           `toml:"default-registry-url,omitempty"`
	DefaultRegistryName string           `toml:"default-registry,omitempty"`
	Experimental        bool             `toml:"experimental,omitempty"`
	TrustedBuilders     []TrustedBuilder `toml:"trusted-builders,omitempty"`
	Registries          []Registry       `toml:"registries,omitempty"`
}

type Registry struct {
	Name string `toml:"name"`
	Type string `toml:"type"`
	URL  string `toml:"url"`
}

type RunImage struct {
	Image   string   `toml:"image"`
	Mirrors []string `toml:"mirrors"`
}

type TrustedBuilder struct {
	Name string `toml:"name"`
}

func DefaultConfigPath() (string, error) {
	home, err := PackHome()
	if err != nil {
		return "", errors.Wrap(err, "getting pack home")
	}
	return filepath.Join(home, "config.toml"), nil
}

func PackHome() (string, error) {
	packHome := os.Getenv("PACK_HOME")
	if packHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "getting user home")
		}
		packHome = filepath.Join(home, ".pack")
	}
	return packHome, nil
}

func Read(path string) (Config, error) {
	cfg := Config{}
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, errors.Wrapf(err, "failed to read config file at path %s", path)
	}

	return cfg, nil
}

func Write(cfg Config, path string) error {
	if err := MkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	return toml.NewEncoder(w).Encode(cfg)
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0777)
}

func SetRunImageMirrors(cfg Config, image string, mirrors []string) Config {
	for i := range cfg.RunImages {
		if cfg.RunImages[i].Image == image {
			cfg.RunImages[i].Mirrors = mirrors
			return cfg
		}
	}
	cfg.RunImages = append(cfg.RunImages, RunImage{Image: image, Mirrors: mirrors})
	return cfg
}

func GetRegistry(cfg Config, registryName string) (Registry, error) {
	if registryName == "" {
		registryName = cfg.DefaultRegistryName
	}
	if registryName != "" {
		for _, registry := range cfg.Registries {
			if registry.Name == registryName {
				return registry, nil
			}
		}
		return Registry{}, errors.Errorf("registry %s is not defined in your config file", style.Symbol(registryName))
	}

	return Registry{
		"official",
		"github",
		"https://github.com/buildpacks/registry-index",
	}, nil
}
