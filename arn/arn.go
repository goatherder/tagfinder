package arn

// Parse Amazon Resource Names.  The 3 different types:
// arn:partition:service:region:account-id:resource-id
// arn:partition:service:region:account-id:resource-type/resource-id
// arn:partition:service:region:account-id:resource-type:resource-id

import (
	"fmt"
	"strings"
)

// ARN is a struct representation of an Amazon Resource Name
type ARN struct {
	Partition         string
	Service           string
	Region            string
	AccountID         string
	ResourceID        string
	ResourceType      string
	ResourceDelimeter string
}

// NewARNFromString returns an ARN struct from the given string
func NewARNFromString(arnString string) (*ARN, error) {
	components := strings.SplitN(arnString, ":", 6)
	if len(components) != 6 {
		return nil, fmt.Errorf("unexpected number of components in arn - expected min of 6 - got: %d", len(components))
	}
	// the last component is the resource-d, or the resource-type/resource-id or resourece-type:resource-id
	resourceText := components[5]
	delim := ":"
	resourceComponents := strings.SplitN(resourceText, delim, 2)
	if len(resourceComponents) == 1 {
		delim = "/"
		resourceComponents = strings.SplitN(resourceText, delim, 2)
	}

	arn := &ARN{
		Partition: components[1],
		Service:   components[2],
		Region:    components[3],
		AccountID: components[4],
	}

	if len(resourceComponents) == 1 {
		arn.ResourceID = resourceComponents[0]
	} else {
		arn.ResourceDelimeter = delim
		arn.ResourceType = resourceComponents[0]
		arn.ResourceID = resourceComponents[1]
	}

	return arn, nil
}

func (a *ARN) String() string {
	if a.ResourceDelimeter != "" {
		return fmt.Sprintf("arn:%s:%s:%s:%s:%s%s%s", a.Partition, a.Service, a.Region, a.AccountID, a.ResourceType, a.ResourceDelimeter, a.ResourceID)
	}
	return fmt.Sprintf("arn:%s:%s:%s:%s:%s", a.Partition, a.Service, a.Region, a.AccountID, a.ResourceID)
}
