package events

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type EventTreeNodeTemplate struct {
	ctx   context.Context
	id    string
	path  string
	event *Event
}

type EventTreeBuilder struct {
}

func (builder *EventTreeBuilder) NewNodeTemplate(
	ctx context.Context,
	event *Event,
) (context.Context, *EventTreeNodeTemplate) {
	parentNodePath := builder.extractNodePathFromContext(ctx)

	newNodeId := uuid.NewString()
	newNodePath := builder.joinNodePaths(parentNodePath, newNodeId)
	newNodeCtx := context.WithValue(ctx, CtxValEventTreeNodePath, newNodePath)
	newNodeTemplate := &EventTreeNodeTemplate{
		ctx:   newNodeCtx,
		id:    newNodeId,
		path:  newNodePath,
		event: event,
	}

	return newNodeCtx, newNodeTemplate
}

func (builder *EventTreeBuilder) NewNode(template *EventTreeNodeTemplate) *EventTreeNode {
	return &EventTreeNode{
		ctx:    template.ctx,
		event:  template.event,
		id:     template.id,
		path:   template.path,
		childs: make(map[string]*EventTreeNode),
	}
}

func (builder *EventTreeBuilder) extractNodePathFromContext(ctx context.Context) string {
	eventPath, _ := ctx.Value(CtxValEventTreeNodePath).(string)
	return eventPath
}

func (builder *EventTreeBuilder) joinNodePaths(parts ...string) string {
	filtered := []string{}
	for i := range parts {
		if len(parts[i]) > 0 {
			filtered = append(filtered, parts[i])
		}
	}
	return strings.Join(filtered, ".")
}
