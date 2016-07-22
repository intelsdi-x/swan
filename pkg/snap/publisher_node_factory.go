package snap

import (
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/pkg/errors"
)

type PublisherNode int

const (
	CassandraPublisherNode PublisherNode = iota
	FilePublisherNode
)

func CreatePublishNode(publisherNode PublisherNode) (*wmap.PublishWorkflowMapNode, error) {
	return CreatePublishNodeWithConfig(publisherNode, nil)
}

func CreatePublishNodeWithConfig(publisherNode PublisherNode, configMap map[string]interface{}) (*wmap.PublishWorkflowMapNode, error) {
	var publisher *wmap.PublishWorkflowMapNode
	switch publisherNode {
	case CassandraPublisherNode:
		publisher = wmap.NewPublishNode("cassandra", 2)
	case FilePublisherNode:
		publisher = wmap.NewPublishNode("file", 3)
	}

	if publisher == nil {
		return nil, errors.New("Failed to create Publish Node for plugin: cassandra, version: 2")
	}

	for key, value := range configMap {
		publisher.AddConfigItem(key, value)
	}

	return publisher, nil
}
