/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"

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
		Name          string   `yaml:"name"`
		Namespace     string   `yaml:"namespace"`
		Authors       []string `yaml:"authors"`
		Description   string   `yaml:"description"`
		License       []string `yaml:"license"`
		LicenseFile   string   `yaml:"license_file"`
		Tags          []string `yaml:"tags"`
		Repository    string   `yaml:"repository"`
		Documentation string   `yaml:"documentation"`
		Homepage      string   `yaml:"homepage"`
		Issues        string   `yaml:"issues"`
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

		// meta ACCUMULATES ACROSS THE FILES OF ONE COLLECTION, SO DROP ANY
		// OPTIONAL FIELDS A PREVIOUS FILE SET BEFORE APPLYING THIS ONE'S
		for field := range GalaxyDefaults {
			delete(meta, field)
		}

		// ONLY DECLARED FIELDS ARE SET SO GalaxyDefaults CAN FILL THE REST
		setMetaString(meta, "description", config.Meta.Description)
		setMetaString(meta, "license_file", config.Meta.LicenseFile)
		setMetaString(meta, "repository", config.Meta.Repository)
		setMetaString(meta, "documentation", config.Meta.Documentation)
		setMetaString(meta, "homepage", config.Meta.Homepage)
		setMetaString(meta, "issues", config.Meta.Issues)
		setMetaList(meta, "authors", config.Meta.Authors)
		setMetaList(meta, "license", config.Meta.License)
		setMetaList(meta, "tags", config.Meta.Tags)
	}

	// ADD REQUIREMENTS TO THE REQUIREMENTS MAP
	if config.Requirements != "" {
		requirements["requirements"] = config.Requirements
	}

	return playbooks, vars, modules, templates, meta, requirements
}

// ENCODES A GALAXY VALUE AS A YAML FLOW SCALAR OR SEQUENCE.
// JSON IS A SUBSET OF YAML 1.2, SO MARSHALLING HANDLES QUOTING/ESCAPING.
// HTML ESCAPING IS OFF SO AN AUTHOR'S <email> STAYS READABLE IN galaxy.yml
func encodeGalaxyValue(key string, value interface{}) (string, bool) {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(value); err != nil {
		log.Printf("Error encoding %s %v: %v", key, value, err)
		return "", false
	}

	// Encode APPENDS A NEWLINE THAT WOULD BREAK THE SINGLE-LINE TEMPLATE SLOT
	return strings.TrimRight(buf.String(), "\n"), true
}

// STORES A NON-EMPTY SCALAR ON THE META MAP
func setMetaString(meta map[string]string, key, value string) {
	if value == "" {
		return
	}

	if encoded, ok := encodeGalaxyValue(key, value); ok {
		meta[key] = encoded
	}
}

// STORES A NON-EMPTY LIST ON THE META MAP
func setMetaList(meta map[string]string, key string, values []string) {
	if len(values) == 0 {
		return
	}

	if encoded, ok := encodeGalaxyValue(key, values); ok {
		meta[key] = encoded
	}
}
