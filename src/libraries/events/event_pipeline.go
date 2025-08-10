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
	p.Next(event, builder)
}

func (p *BaseEventProcessor) Next(event *Event, builder *EventBuilder) {
	if p.next != nil {
		p.next.Process(event, builder)
	}
}

/* EventPipeLine */

type EventPipeline struct {
	processors []EventProcessor
	tail int
}

func (p *EventPipeline) Push(processor EventProcessor) {
	p.processors = append(p.processors, processor)
	if len(p.processors) == 1 {
		return
	} else {
		p.processors[p.tail].SetNext(processor)
		p.tail++ 
	}
}

func (p *EventPipeline) Process(event *Event, builder *EventBuilder) {
	if len(p.processors) > 0 {
		p.processors[0].Process(event, builder)
	}
}
