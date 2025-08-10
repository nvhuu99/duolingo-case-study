package pipelines

import (
	events "duolingo/libraries/events/facade"
	"duolingo/libraries/restful"
	"fmt"
)

type ReceiveRequest struct {
	restful.BasePipeline
}

func (r *ReceiveRequest) Handle(req *restful.Request, res *restful.Response) {
	builder := restful.NewRequestBuilder(req)
	requestCtx := builder.GetRequestContext()

	evt := events.Start(
		requestCtx,
		fmt.Sprintf("restful.%v(%v)", req.Method(), req.URL().Path),
		map[string]any{
			"scheme": req.Scheme(),
			"full_url": req.FullURL(),
			"user_agent": req.UserAgent(),
		},
	)
	defer func() {
		events.End(evt, res.Success(), res.Error(), map[string]any{
			"status_code": res.Status(),
		})
	}()

	builder.SetRequestContext(evt.Context())

	r.Next(req, res)
}
