package options

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	kubeproxyconfig "k8s.io/kubernetes/pkg/proxy/apis/config"
	"sigs.k8s.io/yaml"
)

var sharedFields = []string{"bindAddress"}

type MergeConfiguration struct {
	SharedYAML []byte
	InstanceYAML []byte
	SharedJSON []byte
	InstanceJSON []byte
}

func NewMergeConfiguration(shared, instance []byte) *MergeConfiguration {
	return &MergeConfiguration{
		SharedYAML:   shared,
		InstanceYAML: instance,
	}
}

// convertYAML2JSON translates both shared and instance from YAML to JSON format.
func (c *MergeConfiguration) convertYAMLToJSON() error {
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

// PatchSharedConfiguration retain the keys shared between shared and instance, patches and return the list of
// modified ones.
func (c *MergeConfiguration) MergePatchSharedConfiguration(object *kubeproxyconfig.KubeProxyConfiguration) ([]byte, []string, error) {
	var changedFields []string

	// Convert both configurations from YAML to JSON
	err := c.convertYAMLToJSON()
	if err != nil {
		return nil, nil, err
	}

	var patchedShared = c.SharedJSON
	for _, sharedField := range sharedFields {
		patch := []byte(fmt.Sprintf(`{"$retainKeys": ["%s"]}`, sharedField))

		// Creates a patch with the instance fields.
		instancePatch, err := strategicpatch.StrategicMergePatch(c.InstanceJSON, patch, object)
		if err != nil {
			return nil, nil, err
		}

		// Keep track of the changed fields in the instance.
		if len(instancePatch) > 2 {
			changedFields = append(changedFields, sharedField)
		}

		// Apply the patch back to the original data.
		patchedShared, err = strategicpatch.StrategicMergePatch(patchedShared, instancePatch, object)
		if err != nil {
			return nil, nil, err
		}
	}

	c.SharedJSON = patchedShared
	return patchedShared, changedFields, nil
}