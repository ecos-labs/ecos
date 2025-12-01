package init

import (
	"fmt"

	"github.com/ecos-labs/ecos/code/cli/plugins/types"
)

// GetInitPlugin creates and returns the appropriate init plugin based on cloud provider
func GetInitPlugin(cloudProvider string, force bool, outputPath string) (types.InitPlugin, error) {
	switch cloudProvider {
	case "aws":
		return &AWSCURInitPlugin{
			Force:      force,
			OutputPath: outputPath,
		}, nil
	// case "azure":
	// 	return &AzureInitPlugin{}, nil
	// case "gcp":
	// 	return &GCPInitPlugin{}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", cloudProvider)
	}
}
