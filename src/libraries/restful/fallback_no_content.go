package restful

type fallbackNoContent struct {
	BasePipeline
}

func (r *fallbackNoContent) Handle(req *Request, res *Response) {
	if !res.sent {
		res.NoContent()
	}
	r.Next(req, res)
}
