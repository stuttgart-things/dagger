/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package load

// package main

// import (
// 	"fmt"
// 	"io/fs"
// 	"io/ioutil"
// 	"log"
// 	"path/filepath"

// 	"gopkg.in/yaml.v3"
// )

// // Structure for YAML file
// type Config struct {
// 	Playbooks []struct {
// 		Name string `yaml:"name"`
// 		Play string `yaml:"play"`
// 	} `yaml:"playbooks"`
// 	Vars []struct {
// 		Name string `yaml:"name"`
// 		File string `yaml:"file"`
// 	} `yaml:"vars"`
// 	Templates []struct {
// 		Name string `yaml:"name"`
// 		File string `yaml:"file"`
// 	} `yaml:"templates"`
// 	Meta struct {
// 		Name      string `yaml:"name"`
// 		Namespace string `yaml:"namespace"`
// 	} `yaml:",inline"` // Use inline to match top-level fields
// 	Requirements string `yaml:"requirements"` // Field to capture requirements
// }

// func main() {
// 	// Specify the folder path
// 	folderPath := "/home/sthings/projects/stuttgart-things/ansible/collections/container"

// 	// Maps to store playbooks, vars, templates, meta, and requirements
// 	playbooks := make(map[string]string)
// 	vars := make(map[string]string)
// 	templates := make(map[string]string)
// 	meta := make(map[string]string)
// 	requirements := make(map[string]string)

// 	// Read all YAML files from the folder
// 	err := filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		// Process only YAML files
// 		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
// 			// Read and process the file
// 			processFile(path, playbooks, vars, templates, meta, requirements)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		log.Fatalf("Error reading folder: %v", err)
// 	}

// 	// Print the playbooks map
// 	fmt.Println("Playbooks:")
// 	for key, value := range playbooks {
// 		fmt.Printf("Key: %s\nContent:\n%s\n\n", key, value)
// 	}

// 	// Print the vars map
// 	fmt.Println("Vars:")
// 	for key, value := range vars {
// 		fmt.Printf("Key: %s\nContent:\n%s\n\n", key, value)
// 	}

// 	// Print the templates map
// 	fmt.Println("Templates:")
// 	for key, value := range templates {
// 		fmt.Printf("Key: %s\nContent:\n%s\n\n", key, value)
// 	}

// 	// Print the meta map
// 	fmt.Println("Meta:")
// 	for key, value := range meta {
// 		fmt.Printf("Key: %s\nValue: %s\n", key, value)
// 	}

// 	// Print the requirements map
// 	fmt.Println("Requirements:")
// 	for key, value := range requirements {
// 		fmt.Printf("Key: %s\nContent:\n%s\n\n", key, value)
// 	}
// }

// func processFile(filePath string, playbooks, vars, templates, meta, requirements map[string]string) {
// 	// Read file content
// 	data, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		log.Printf("Error reading file %s: %v", filePath, err)
// 		return
// 	}

// 	// Parse YAML content
// 	var config Config
// 	err = yaml.Unmarshal(data, &config)
// 	if err != nil {
// 		log.Printf("Error parsing YAML file %s: %v", filePath, err)
// 		return
// 	}

// 	// Add playbooks to the playbooks map
// 	for _, playbook := range config.Playbooks {
// 		playbooks[playbook.Name] = playbook.Play
// 	}

// 	// Add vars to the vars map
// 	for _, variable := range config.Vars {
// 		vars[variable.Name] = variable.File
// 	}

// 	// Add templates to the templates map
// 	for _, template := range config.Templates {
// 		templates[template.Name] = template.File
// 	}

// 	// Add meta information if both keys are present
// 	if config.Meta.Name != "" && config.Meta.Namespace != "" {
// 		meta["name"] = config.Meta.Name
// 		meta["namespace"] = config.Meta.Namespace
// 	}

// 	// Add requirements to the requirements map
// 	if config.Requirements != "" {
// 		requirements[filePath] = config.Requirements
// 	}
// }
