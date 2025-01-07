/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Playbooks []struct {
		Name string `yaml:"name"`
		Play string `yaml:"play"`
	} `yaml:"playbooks"`
	Vars []struct {
		Name string `yaml:"name"`
		File string `yaml:"file"`
	} `yaml:"vars"`
	Templates []struct {
		Name string `yaml:"name"`
		File string `yaml:"file"`
	} `yaml:"templates"`
	Modules []struct {
		Name string `yaml:"name"`
		File string `yaml:"file"`
	} `yaml:"modules"`
	Meta struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		// Authors   []string `yaml:"authors"`
	} `yaml:",inline"` // Use inline to match top-level fields
	Requirements string `yaml:"requirements"` // Field to capture requirements
}

// ProcessFile reads and processes the YAML file
func ProcessCollectionFile(data []byte, playbooks, vars, modules, templates, meta, requirements map[string]string) (map[string]string, map[string]string, map[string]string, map[string]string, map[string]string, map[string]string) {

	// PARSE YAML CONTENT
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error parsing YAML file %s: %v", data, err)
	}

	// ADD PLAYBOOKS TO THE PLAYBOOKS MAP
	for _, playbook := range config.Playbooks {
		playbooks[playbook.Name] = playbook.Play
	}

	// ADD VARS TO THE VARS MAP
	for _, variable := range config.Vars {
		vars[variable.Name] = variable.File
	}

	// ADD TEMPLATES TO THE TEMPLATES MAP
	for _, template := range config.Templates {
		templates[template.Name] = template.File
	}

	// ADD MODULES TO THE MODULES MAP
	for _, module := range config.Modules {
		modules[module.Name] = module.File
	}

	// ADD META INFORMATION IF BOTH KEYS ARE PRESENT
	if config.Meta.Name != "" && config.Meta.Namespace != "" {
		meta["name"] = config.Meta.Name
		meta["namespace"] = config.Meta.Namespace
		// meta["authors"] = config.Meta.Authors[0]
	}

	// ADD REQUIREMENTS TO THE REQUIREMENTS MAP
	if config.Requirements != "" {
		requirements["requirements"] = config.Requirements
	}

	return playbooks, vars, modules, templates, meta, requirements
}
