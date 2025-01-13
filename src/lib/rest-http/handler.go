package resthttp

type Handler struct {
	Next *Handler
	Handle func(req *Request, res *Response)
}
