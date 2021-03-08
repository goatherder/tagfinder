package tags_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	resapi "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	resapiiface "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/goatherder/tagfinder/tags"
	log "github.com/sirupsen/logrus"
)

var (
	testGetResourcesResponse *resapi.GetResourcesOutput = &resapi.GetResourcesOutput{
		PaginationToken: aws.String(""),
		ResourceTagMappingList: []*resapi.ResourceTagMapping{
			{
				ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test/abc123"),
				Tags: []*resapi.Tag{
					{
						Key:   aws.String("tag1"),
						Value: aws.String("value1"),
					},
					{
						Key:   aws.String("tag:test"),
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
						Key:   aws.String("tag:test"),
						Value: aws.String("stuff"),
					},
				},
			},
		},
	}
)

type MockResourceGroupsTaggingAPIClient struct {
	resapiiface.ResourceGroupsTaggingAPIAPI
}

func (m *MockResourceGroupsTaggingAPIClient) GetResources(input *resapi.GetResourcesInput) (*resapi.GetResourcesOutput, error) {
	return testGetResourcesResponse, nil
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
	for idx, r := range resourceList {
		t.Logf("resource[%d]: %+v", idx, r)
	}
}
