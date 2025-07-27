package events

import (
	"context"
	"strings"
	"sync"
	"time"
)

type EventContextValue string

const (
	CtxValEventTreeNodePath EventContextValue = "event_tree.node_path"
)

type EventTreeNode struct {
	// Node data
	ctx       context.Context
	event     *Event
	endedFlag bool

	// Node properties
	id     string
	path   string
	isRoot bool
	parent *EventTreeNode
	childs map[string]*EventTreeNode
}

type EventTreeRoot struct {
	treeMutex sync.Mutex
	rootNode  *EventTreeNode
	builder   *EventTreeBuilder
}

func NewEventTreeRoot() *EventTreeRoot {
	return &EventTreeRoot{
		builder: &EventTreeBuilder{},
		rootNode: &EventTreeNode{
			isRoot: true,
			childs: make(map[string]*EventTreeNode),
		},
	}
}

func (root *EventTreeRoot) InsertNode(newNode *EventTreeNode) {
	root.treeMutex.Lock()
	defer root.treeMutex.Unlock()

	parts := strings.Split(newNode.path, ".")
	parentPath := strings.Join(parts[0:len(parts)-1], ".")
	parentNode := root.find(parentPath)

	if parentNode == nil {
		newNode.parent = root.rootNode
		root.rootNode.childs[newNode.id] = newNode
	} else {
		newNode.parent = parentNode
		parentNode.childs[newNode.id] = newNode
	}
}

func (root *EventTreeRoot) find(path string) *EventTreeNode {
	parts := strings.Split(path, ".")
	travel := root.rootNode
	for i := range parts {
		for id, node := range travel.childs {
			if parts[i] != id {
				continue
			}
			if i == len(parts)-1 {
				return node
			}
			travel = node
			break
		}
	}
	return nil
}

func (root *EventTreeRoot) FindNodeFromContextAndFlagEventEnded(
	ctx context.Context,
	endedAt time.Time,
) {
	root.treeMutex.Lock()
	defer root.treeMutex.Unlock()

	path := root.builder.extractNodePathFromContext(ctx)
	node := root.find(path)

	node.endedFlag = true
	node.event.endedAt = endedAt
}

func (root *EventTreeRoot) ExtractAllEndedEvents() []*Event {
	root.treeMutex.Lock()
	defer root.treeMutex.Unlock()

	endedNodes := []*EventTreeNode{}
	for _, rootChild := range root.rootNode.childs {
		if root.hasAllEventsInPathEnded(rootChild) {
			endedNodes = append(endedNodes, rootChild)
		}
	}

	for i := range endedNodes {
		root.ensureEventEndTimeMatchesLatestChild(endedNodes[i])
	}

	endedEvents := []*Event{}
	for i := range endedNodes {
		root.travel(endedNodes[i], func(node *EventTreeNode) {
			endedEvents = append(endedEvents, node.event)
		})
	}

	for i := range endedNodes {
		delete(root.rootNode.childs, endedNodes[i].id)
	}

	return endedEvents
}

func (root *EventTreeRoot) hasAllEventsInPathEnded(node *EventTreeNode) bool {
	for _, child := range node.childs {
		if ended := root.hasAllEventsInPathEnded(child); !ended {
			return false
		}
	}
	return node.endedFlag
}

func (root *EventTreeRoot) ensureEventEndTimeMatchesLatestChild(node *EventTreeNode) {
	if len(node.childs) == 0 {
		return
	}

	for _, child := range node.childs {
		root.ensureEventEndTimeMatchesLatestChild(child)
	}

	for _, child := range node.childs {
		if node.event.endedAt.Before(child.event.endedAt) {
			node.event.endedAt = child.event.endedAt
		}
	}
}

func (root *EventTreeRoot) travel(node *EventTreeNode, visitor func(*EventTreeNode)) {
	if node == nil {
		return
	}
	visitor(node)
	for _, node := range node.childs {
		root.travel(node, visitor)
	}
}
