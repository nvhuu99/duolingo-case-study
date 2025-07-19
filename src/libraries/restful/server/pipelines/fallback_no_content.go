package pipelines

import "duolingo/libraries/restful"

type FallbackNoContent struct {
	restful.BasePipeline
}

func (r *FallbackNoContent) Handle(req *restful.Request, res *restful.Response) {
	res.NoContent()
	r.Next(req, res)
}
