package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func NewConfig(file string) (*Config, error) {
	yfile, err := ioutil.ReadFile(file)
	if err != nil {
		return &Config{}, fmt.Errorf("error parsing values file: %w", err)
	}

	config := &Config{}

	err = yaml.Unmarshal(yfile, config)
	if err != nil {
		return &Config{}, fmt.Errorf("error unmarshal'ing file: %w", err)
	}

	if config.Org == "" {
		return &Config{}, fmt.Errorf("error - missing Pulumi Org")
	}

	if config.Aws.Region == "" {
		config.Aws.Region = "us-east-2"
	}

	return config, nil
}

type Config struct {
	Version   string            `yaml:"version"`
	Name      string            `yaml:"name"`
	Slug      string            `yaml:"slug"`
	Org       string            `yaml:"org"`
	Owner     string            `yaml:"owner"`
	Slack     string            `yaml:"slack,omitempty"`
	Namespace string            `yaml:"namespace"`
	Deploy    map[string]Deploy `yaml:"deploy"`
	Aws       AwsResources      `yaml:"aws,omitempty"`
}

type AwsResources struct {
	IamRole  bool     `yaml:"iamRole,omitempty"`
	Region   string   `yaml:"region,omitempty"`
	S3Bucket S3Bucket `yaml:"s3Bucket,omitempty"`
	SQS      []SQS    `yaml:"sqs,omitempty"`
}

type SQS struct {
	Name string `yaml:"name,omitempty"`
}

type S3Bucket struct {
	Create  bool   `yaml:"create"`
	Name    string `yaml:"name,omitempty"`
	Encrypt bool   `yaml:"encrtyp,omitempty"`
}

type Deploy struct {
	EnvVars   []EnvVars `yaml:"env"`
	Ports     []Ports   `yaml:"ports"`
	Service   []Service `yaml:"service"`
	Ingress   []Ingress `yaml:"ingress"`
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	Requests CpuMem `yaml:"requests"`
	Limits   CpuMem `yaml:"limits"`
}

type CpuMem struct {
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type Ingress struct {
	Host     string `yaml:"host"`
	IsPublic bool   `yaml:"public"`
}

type Service struct {
	Name  string  `yaml:"name"`
	Ports []Ports `yaml:"ports"`
}

type Ports struct {
	Name   string `yaml:"name"`
	Number int    `yaml:"number"`
}

type EnvVars struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Application struct {
	Container Container `yaml:"container"`
}

type Container struct {
	Repository string `yaml:"repository"`
	PullPolicy string `yaml:"pullPolicy,omitempty"`
}
