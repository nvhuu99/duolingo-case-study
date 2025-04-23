package writer

type Writable struct {
	URI      string `json:"uri"`
	Rotation string `json:"rotation"`
	Content  []byte `json:"content"`
}
