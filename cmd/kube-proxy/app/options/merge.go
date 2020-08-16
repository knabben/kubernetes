package options

import (
	"github.com/evanphx/json-patch"
	"sigs.k8s.io/yaml"
)

type Configuration struct {
	SharedYAML []byte
	InstanceYAML []byte
	SharedJSON []byte
	InstanceJSON []byte
}

func NewMergeConfiguration(shared, instance []byte) *Configuration {
	return &Configuration{
		SharedYAML:   shared,
		InstanceYAML: instance,
	}
}

// convertYAML2JSON translates both shared and instance from YAML to JSON format.
func (c *Configuration) convertYAML2JSON() error {
	var err error

	c.SharedJSON, err = yaml.YAMLToJSON(c.SharedYAML)
	if err != nil {
		return err
	}

	c.InstanceJSON, err = yaml.YAMLToJSON(c.InstanceYAML)
	if err != nil {
		return err
	}

	return nil
}

func (c *Configuration) PatchShared() ([]byte, error) {
	// Convert both YAML to JSON patch
	err := c.convertYAML2JSON()
	if err != nil {
		return nil, err
	}

	// Cleanup the instance JSON, keeping only the data fields.
	cleanupPatch := `[{"op": "remove", "path": "/kind"}, {"op": "remove", "path": "/apiVersion"}]`
	patchJSON, err := jsonpatch.DecodePatch([]byte(cleanupPatch))
	if err != nil {
		return nil, err
	}

	// Apply the patch in the instance JSON
	instanceClean, err := patchJSON.Apply(c.InstanceJSON)
	if err != nil {
		return nil, err
	}

	// Merge the shared JSON into the instance keeping Instance specific values
	patchedShared, err := jsonpatch.MergeMergePatches(c.SharedJSON, instanceClean)
	if err != nil {
		return nil, err
	}

	c.SharedJSON = patchedShared

	return patchedShared, nil
}