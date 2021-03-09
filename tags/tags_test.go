package tags_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	resapi "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	resapiiface "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/goatherder/tagfinder/arn"
	"github.com/goatherder/tagfinder/tags"
	log "github.com/sirupsen/logrus"
)

var (
	testResourceTagMappings []*resapi.ResourceTagMapping = []*resapi.ResourceTagMapping{
		{
			ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test/abc123"),
			Tags: []*resapi.Tag{
				{
					Key:   aws.String("name"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("thing:test"),
					Value: aws.String("thing"),
				},
			},
		},
		{
			ResourceARN: aws.String("arn:aws:rds:us-east-1:123456789012:db:test"),
			Tags: []*resapi.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String("test"),
				},
				{
					Key:   aws.String("thing:test"),
					Value: aws.String("stuff"),
				},
			},
		},
	}
	testGetResourcesResponse *resapi.GetResourcesOutput = &resapi.GetResourcesOutput{
		PaginationToken:        aws.String(""),
		ResourceTagMappingList: testResourceTagMappings,
	}
)

func ExpectedResources() *resapi.GetResourcesOutput {
	return testGetResourcesResponse
}

type MockResourceGroupsTaggingAPIClient struct {
	resapiiface.ResourceGroupsTaggingAPIAPI
}

func (m *MockResourceGroupsTaggingAPIClient) GetResources(input *resapi.GetResourcesInput) (*resapi.GetResourcesOutput, error) {

	if input.TagFilters == nil && input.ResourceTypeFilters == nil {
		return testGetResourcesResponse, nil
	}

	// apply some filters
	tagMappingList := make([]*resapi.ResourceTagMapping, 0)

	for _, r := range testResourceTagMappings {
		add := false
		if input.TagFilters != nil {
			for _, tf := range input.TagFilters {
				for _, tag := range r.Tags {
					// FIXME: assume only a single value
					if unrefstr(tag.Key) == unrefstr(tf.Key) && unrefstr(tag.Value) == unrefstr(tf.Values[0]) {
						add = true
						break
					}
				}
			}
		}
		if !add && input.ResourceTypeFilters != nil {
			arn, err := arn.NewARNFromString(unrefstr(r.ResourceARN))
			if err != nil {
				return nil, fmt.Errorf("error parsing test resource arn: %v", err)
			}

			for _, rtf := range input.ResourceTypeFilters {
				if unrefstr(rtf) == arn.Service {
					add = true
					break
				}
			}
		}
		if add {
			tagMappingList = append(tagMappingList, r)
		}
	}

	resp := &resapi.GetResourcesOutput{
		PaginationToken:        aws.String(""),
		ResourceTagMappingList: tagMappingList,
	}

	return resp, nil
}

var (
	tagsclient *tags.Client
	l          *log.Logger
)

func setup(t *testing.T) {
	var err error
	mc := &MockResourceGroupsTaggingAPIClient{}
	l := log.New()
	l.SetLevel(log.DebugLevel)
	tagsclient, err = tags.New(tags.WithResourceGroupsTaggingAPIClient(mc), tags.WithLogger(l))
	if err != nil {
		t.Fatalf("Failed initializing mock tags client: %v", err)
	}
}

func TestGetResources(t *testing.T) {
	setup(t)
	resourceList, err := tagsclient.GetResources()
	if err != nil {
		t.Fatalf("Error fetching resource list: %v", err)
	}
	if resourceList == nil {
		t.Errorf("resourceList was nil")
	}
	if len(resourceList) != len(testGetResourcesResponse.ResourceTagMappingList) {
		t.Errorf("length of response (%d) didn't match expected (%d)", len(resourceList), len(testGetResourcesResponse.ResourceTagMappingList))
	}
	for idx, r := range resourceList {
		t.Logf("resource[%d]: %+v", idx, r)
	}
}

func TestGetResourcesWithTags(t *testing.T) {
	setup(t)
	tagf := map[string]string{"Name": "test"}
	expectedResArn := "arn:aws:rds:us-east-1:123456789012:db:test"
	resourceList, err := tagsclient.GetResources(tags.WithTags(tagf))
	if err != nil {
		t.Fatalf("Error fetching resource list: %v", err)
	}
	if len(resourceList) != 1 {
		t.Fatalf("len(resourceList) (%d) != 1", len(resourceList))
	}
	if resourceList[0].ARN != expectedResArn {
		t.Errorf("resource arn (%s) didn't match expected (%s)", resourceList[0].ARN, expectedResArn)
	}
}

func TestGetResourcesWithType(t *testing.T) {
	setup(t)
	rtypes := []string{
		"elasticloadbalancing",
	}
	expectedResArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test/abc123"
	resourceList, err := tagsclient.GetResources(tags.WithResourceTypes(rtypes))
	if err != nil {
		t.Fatalf("Error fetching resource list: %v", err)
	}
	if len(resourceList) != 1 {
		t.Fatalf("len(resourceList) (%d) != 1", len(resourceList))
	}
	if resourceList[0].ARN != expectedResArn {
		t.Errorf("resource arn (%s) didn't match expected (%s)", resourceList[0].ARN, expectedResArn)
	}
}

func unrefstr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
