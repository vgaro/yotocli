package config

import (
        "os"
        "path/filepath"

        "gopkg.in/yaml.v3"
)

type IconRecord struct {
        ID   string   `yaml:"id"`
        Name string   `yaml:"name"`
        Tags []string `yaml:"tags,omitempty"`
}

type IconsConfig struct {
        Icons []IconRecord `yaml:"icons"`
}

func LoadIcons() (*IconsConfig, error) {
        home, _ := os.UserHomeDir()
        path := filepath.Join(home, ".config", "yotocli", "icons.yaml")

        data, err := os.ReadFile(path)
        if err != nil {
                if os.IsNotExist(err) {
                        return &IconsConfig{Icons: []IconRecord{}}, nil
                }
                return nil, err
        }

        var config IconsConfig
        if err := yaml.Unmarshal(data, &config); err != nil {
                return nil, err
        }
        return &config, nil
}

func SaveIcons(config *IconsConfig) error {
        home, _ := os.UserHomeDir()
        configPath := filepath.Join(home, ".config", "yotocli")
        path := filepath.Join(configPath, "icons.yaml")

        if err := os.MkdirAll(configPath, 0755); err != nil {
                return err
        }

        data, err := yaml.Marshal(config)
        if err != nil {
                return err
        }

        return os.WriteFile(path, data, 0644)
}

func AddIconRecord(id, name string, tags []string) error {
        config, err := LoadIcons()
        if err != nil {
                return err
        }

        // Check if already exists
        for i, icon := range config.Icons {
                if icon.ID == id {
                        config.Icons[i].Name = name
                        config.Icons[i].Tags = tags
                        return SaveIcons(config)
                }
        }

        config.Icons = append(config.Icons, IconRecord{
                ID:   id,
                Name: name,
                Tags: tags,
        })
        return SaveIcons(config)
}
