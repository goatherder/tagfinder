// Package tags wraps AWS Resource Groups Tagging API in a basic interface that provides calls specific to our needs
package tags

// Note: from aws docs: Allowed characters can vary by AWS service. For information about what characters you can use
//   to tag resources in a particular AWS service, see its documentation. In general, allowed characters in tags are
//   letters, numbers, spaces representable in UTF-8, and the following characters: _ . : / = + - @ .

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	resapi "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	resapiiface "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	log "github.com/sirupsen/logrus"
)

// Client is our own parameter store client
type Client struct {
	resClient resapiiface.ResourceGroupsTaggingAPIAPI
	l         *log.Logger
}

// ClientOptions provides the ability to override client settings during initialization
type ClientOptions func(*Client)

// WithResourceGroupsTaggingAPIClient allows providing the AWS Resource Groups Tagging API client instead of initializing one
func WithResourceGroupsTaggingAPIClient(resc resapiiface.ResourceGroupsTaggingAPIAPI) ClientOptions {
	return func(c *Client) {
		c.resClient = resc
	}
}

// WithLogger allows setting the logger
func WithLogger(l *log.Logger) ClientOptions {
	return func(c *Client) {
		c.l = l
	}
}

// New returns a new Client instance
func New(opts ...ClientOptions) (*Client, error) {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	if c.l == nil {
		c.l = log.New()
		c.l.SetFormatter(&log.JSONFormatter{})
	}

	l := c.l.WithFields(log.Fields{"package": "tags", "method": "New"})

	if c.resClient == nil {
		l.Debugf("initializing aws session")
		cfg := &aws.Config{}

		sess := session.Must(session.NewSession())
		c.resClient = resapi.New(sess, cfg)
	}

	return c, nil
}

// GetResourcesFilterOpts allows chaining optional additions to the GetResources filter
type GetResourcesFilterOpts func(*resapi.GetResourcesInput)

// WithTags modifies the request filter with the given list of tags (map[string]string - key value)
// NOTE: assumes only one value per key
func WithTags(tags map[string]string) GetResourcesFilterOpts {
	return func(r *resapi.GetResourcesInput) {
		tf := make([]*resapi.TagFilter, len(tags))
		idx := 0
		for k, v := range tags {
			tf[idx] = &resapi.TagFilter{
				Key: aws.String(k),
				Values: []*string{
					aws.String(v),
				},
			}
			idx++
		}
		r.TagFilters = tf
	}
}

// WithResourceTypes modifies the request filter with the given resource type list.  These are the resource types as appears
// in the AWS ARN - eg: ec2, s3, etc...
func WithResourceTypes(resources []string) GetResourcesFilterOpts {
	return func(r *resapi.GetResourcesInput) {
		rtf := make([]*string, len(resources))
		for idx, s := range resources {
			rtf[idx] = aws.String(s)
		}
		r.ResourceTypeFilters = rtf
	}
}

// Resource is a basic struct containing details on resources we're interested in
type Resource struct {
	ARN  string
	Tags map[string]string
}

// GetResources returns all AWS resources given the specified filters
func (c *Client) GetResources(filteropts ...GetResourcesFilterOpts) ([]*Resource, error) {
	l := c.l.WithFields(log.Fields{"package": "tags", "method": "GetResources"})
	input := &resapi.GetResourcesInput{}

	for _, filteropt := range filteropts {
		filteropt(input)
	}

	resources := make([]*Resource, 0)
	pageCount := 0

	for {
		output, err := c.resClient.GetResources(input)
		if err != nil {
			l.Errorf("There was an error during the call to GetResources: %v", err)
			return nil, err
		}

		if output == nil {
			l.Errorf("response from GetResources was nil")
			return nil, fmt.Errorf("aws response was nil")
		}

		l.Infof("processing resources page %d", pageCount)

		for _, resource := range output.ResourceTagMappingList {
			tags := make(map[string]string, 0)
			for _, t := range resource.Tags {
				if _, ok := tags[*t.Key]; ok {
					l.Warnf("resource %s - dup tag key %s - clobbering! (old val: %s, new val: %s)", *resource.ResourceARN, *t.Key, tags[*t.Key], *t.Value)
				}
				if t.Key == nil {
					return nil, fmt.Errorf("tag key was nil")
				}
				if t.Value == nil {
					return nil, fmt.Errorf("tag value was nil")
				}
				tags[*t.Key] = *t.Value
			}
			newRes := &Resource{
				ARN:  *resource.ResourceARN,
				Tags: tags,
			}
			resources = append(resources, newRes)
		}

		// if no pagination token, then we're done
		if output.PaginationToken == nil {
			l.Debugf("paginationtoken was nil - exiting loop")
			break
		}
		if *output.PaginationToken == "" {
			l.Debugf("paginationtoken was an empty string - exiting loop")
			break
		}

		pageCount++
		input.PaginationToken = output.PaginationToken
	}

	return resources, nil
}
