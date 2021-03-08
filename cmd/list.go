package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/goatherder/tagfinder/tags"
	"github.com/spf13/cobra"
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
	tags, err := tags.New(tags.WithLogger(l))
	if err != nil {
		l.Errorf("error intializing tags client: %v", err)
	}
	resources, err := tags.GetResources()
	if err != nil {
		l.Errorf("error listing resources: %v", err)
	}
	for idx, r := range resources {
		fmt.Printf("resource[%d]: %s: tags: %v\n", idx, r.ARN, r.Tags)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
