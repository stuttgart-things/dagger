/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package templates

import (
	"bytes"
	"html/template"
	"log"
)

func RenderTemplate(templateData string, data map[string]interface{}) (renderTemplate string) {

	// PARSE THE TEMPLATEs
	t, err := template.New("dagger-xplane").Parse(templateData)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// EXECUTE THE TEMPLATE AND WRITE THE OUTPUT TO A BUFFER
	var output bytes.Buffer
	err = t.Execute(&output, data)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	renderTemplate = output.String()

	return renderTemplate
}
