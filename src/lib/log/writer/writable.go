package writer

type Writable struct {
	Namespace string
	Prefix    string
	Extension string
	Content   []byte
}
