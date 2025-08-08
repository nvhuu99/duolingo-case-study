package events

/* EventProcessor */

type EventProcessor interface {
	SetNext(EventProcessor)
	Process(*Event, *EventBuilder)
	Next(*Event, *EventBuilder)
}

type EventDecorator interface {
	Decorate(*Event, *EventBuilder)
}

type EventFinalizer interface {
	Finalize(*Event, *EventBuilder)
}

type BaseEventProcessor struct {
	next EventProcessor
	processFunc func(*Event, *EventBuilder)
}

func NewBaseEventProcessor(processFunc func(*Event, *EventBuilder)) *BaseEventProcessor {
	return &BaseEventProcessor{
		processFunc: processFunc,
	}
}

func (p *BaseEventProcessor) SetNext(next EventProcessor) {
	p.next = next
}

func (p *BaseEventProcessor) Process(event *Event, builder *EventBuilder) {
	p.processFunc(event, builder)
}

func (p *BaseEventProcessor) Next(event *Event, builder *EventBuilder) {
	if p.next != nil {
		p.next.Process(event, builder)
	}
}

/* EventPipeLine */

type EventPipeline struct {
	stages []EventProcessor
}

func (p *EventPipeline) Push(stages ...EventProcessor) {
	p.stages = append(p.stages, stages...)
}

func (p *EventPipeline) Process(event *Event, builder *EventBuilder) {
	if len(p.stages) > 0 {
		p.stages[0].Process(event, builder)
	}
}
