// pkg/config/docker.go
package config

type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
	Volumes  map[string]interface{} `yaml:"volumes"`
}

type Service struct {
	Image         string `yaml:"image"`
	ContainerName string `yaml:"container_name"`
}
