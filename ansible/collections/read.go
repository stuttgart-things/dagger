/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package collections

import (
	"fmt"
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
	Meta struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:",inline"` // Use inline to match top-level fields
	Requirements string `yaml:"requirements"` // Field to capture requirements
}

func Test(name string) {
	fmt.Println(name)
}

// func ReadCollectionFiles(folderPath string) (map[string]string, map[string]string, map[string]string, map[string]string, map[string]string) {

// 	// MAPS TO STORE PLAYBOOKS, VARS, TEMPLATES, META, AND REQUIREMENTS
// 	playbooks := make(map[string]string)
// 	vars := make(map[string]string)
// 	templates := make(map[string]string)
// 	meta := make(map[string]string)
// 	requirements := make(map[string]string)

// 	// READ ALL YAML FILES FROM THE FOLDER
// 	err := filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		// PROCESS ONLY YAML FILES
// 		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
// 			// READ AND PROCESS THE FILE
// 			processFile(path, playbooks, vars, templates, meta, requirements)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		log.Fatalf("ERROR READING FOLDER: %v", err)
// 	}

// 	return playbooks, vars, templates, meta, requirements
// }

// ProcessFile reads and processes the YAML file
func ProcessCollectionFile(data []byte, playbooks, vars, templates, meta, requirements map[string]string) (map[string]string, map[string]string, map[string]string, map[string]string, map[string]string) {

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

	// ADD META INFORMATION IF BOTH KEYS ARE PRESENT
	if config.Meta.Name != "" && config.Meta.Namespace != "" {
		meta["name"] = config.Meta.Name
		meta["namespace"] = config.Meta.Namespace
	}

	// ADD REQUIREMENTS TO THE REQUIREMENTS MAP
	if config.Requirements != "" {
		requirements["requirements"] = config.Requirements
	}

	// fmt.Println("PLAYBOOKS", playbooks)
	// fmt.Println("VARS", vars)
	// fmt.Println("TEMPLATES", templates)
	// fmt.Println("META", meta)
	// fmt.Println("REQUIRMENTS", requirements)

	return playbooks, vars, templates, meta, requirements
}
