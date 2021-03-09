package arn_test

import (
	"reflect"
	"testing"

	"github.com/goatherder/tagfinder/arn"
)

func TestARNParse(t *testing.T) {
	tests := []struct {
		ARN      string
		Expected *arn.ARN
	}{
		{
			ARN: "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test/abc123",
			Expected: &arn.ARN{
				Partition:         "aws",
				Service:           "elasticloadbalancing",
				Region:            "us-east-1",
				AccountID:         "123456789012",
				ResourceType:      "loadbalancer",
				ResourceID:        "app/test/abc123",
				ResourceDelimeter: "/",
			},
		},
		{
			ARN: "arn:aws:rds:us-east-1:123456789012:db:testdb",
			Expected: &arn.ARN{
				Partition:         "aws",
				Service:           "rds",
				Region:            "us-east-1",
				AccountID:         "123456789012",
				ResourceType:      "db",
				ResourceID:        "testdb",
				ResourceDelimeter: ":",
			},
		},
		{
			ARN: "arn:aws:s3:::testbucket",
			Expected: &arn.ARN{
				Partition:  "aws",
				Service:    "s3",
				ResourceID: "testbucket",
			},
		},
	}

	for _, test := range tests {
		arn, err := arn.NewARNFromString(test.ARN)
		if err != nil {
			t.Fatalf("error parsing arn: %v", err)
		}
		if !reflect.DeepEqual(arn, test.Expected) {
			t.Errorf("arn (%+v) didn't match expected (%+v) - from string %s", arn, test.Expected, test.ARN)
		}
	}

}
