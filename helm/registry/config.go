/*
Copyright © 2024 Patrick Hermann patrick.hermann@sva.de
*/

package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// FUNCTION TO CREATE DOCKER CONFIG.JSON AND RETURN IT AS A STRING
func CreateDockerConfigJSON(username, password, registry string) (string, error) {
	// ENCODE USERNAME:PASSWORD IN BASE64
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	// CREATE CONFIG.JSON STRUCTURE
	config := map[string]interface{}{
		"auths": map[string]interface{}{
			registry: map[string]string{
				"auth": auth,
			},
		},
	}

	// SERIALIZE CONFIG.JSON TO JSON
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("FAILED TO MARSHAL config.json: %w", err)
	}

	// RETURN THE JSON STRING
	return string(configData), nil
}
