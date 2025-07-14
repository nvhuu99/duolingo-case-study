package restful

type Pipeline interface {
	SetNext(Pipeline)
	Handle(*Request, *Response)
	Next(*Request, *Response)
}

type BasePipeline struct {
	next Pipeline
}

func (p *BasePipeline) SetNext(next Pipeline) {
	p.next = next
}

func (p *BasePipeline) Handle(req *Request, res *Response) {}

func (p *BasePipeline) Next(req *Request, res *Response) {
	if p.next != nil {
		p.next.Handle(req, res)
	}
}
