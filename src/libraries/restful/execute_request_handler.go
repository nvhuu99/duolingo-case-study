package restful

type executeRequestHandler struct {
	BasePipeline
}

func (r *executeRequestHandler) Handle(req *Request, res *Response) {
	defer r.panicHandler(res)
	if req.handler != nil {
		req.handler(req, res)
	}
	r.Next(req, res)
}

func (r *executeRequestHandler) panicHandler(response *Response) {
	if r := recover(); r != nil {
		response.ServerErr("failed to handle request")
	}
}
