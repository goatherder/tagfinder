package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/goatherder/tagfinder/tags"
	"github.com/spf13/cobra"
)

var (
	tagKeyVal    string
	serviceTypes string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list resources",
	Long:  `List AWS resources`,
	Run:   listResources,
}

func listResources(cmd *cobra.Command, args []string) {
	l := log.New()
	t, err := tags.New(tags.WithLogger(l))
	if err != nil {
		l.Errorf("error intializing tags client: %v", err)
	}

	filters := make([]tags.GetResourcesFilterOpts, 0)

	if tagKeyVal != "" {
		keyvals := strings.Split(tagKeyVal, ",")
		if len(keyvals)%2 != 0 {
			fmt.Printf("key val count was not even (you need to specify tags in key/value pairs - comma separator, no spaces")
			return
		}
		keyvalmap := make(map[string]string, len(keyvals)/2)
		for i := 0; i < len(keyvals)/2; i++ {
			keyvalmap[keyvals[i*2]] = keyvals[(i*2)+1]
		}
		fmt.Printf("filtering by tags: %v\n", keyvalmap)
		filters = append(filters, tags.WithTags(keyvalmap))
	}

	if serviceTypes != "" {
		services := strings.Split(serviceTypes, ",")
		fmt.Printf("filtering by services: %v\n", services)
		filters = append(filters, tags.WithResourceTypes(services))
	}

	resources, err := t.GetResources(filters...)
	if err != nil {
		l.Errorf("error listing resources: %v", err)
	}
	for idx, r := range resources {
		fmt.Printf("resource[%d]: %s: tags: %v\n", idx, r.ARN, r.Tags)
	}
}

func init() {
	listCmd.PersistentFlags().StringVarP(&tagKeyVal, "tags", "t", "", "Tag filter")
	listCmd.PersistentFlags().StringVarP(&serviceTypes, "services", "s", "", "Service filter")
	rootCmd.AddCommand(listCmd)
}
