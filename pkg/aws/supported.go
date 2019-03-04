package aws

import (
	"github.com/go-errors/errors"
	"github.com/iflix/awsweeper/pkg/terraform"
)

func SupportedResourceType(resType terraform.ResourceType) bool {
	_, found := deleteIDs[resType]

	return found
}

func getDeleteID(resType terraform.ResourceType) (string, error) {
	deleteID, found := deleteIDs[resType]
	if !found {
		return "", errors.Errorf("no delete ID specified for resource type: %s", resType)
	}
	return deleteID, nil
}
