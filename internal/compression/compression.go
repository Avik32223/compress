package compression

type Engine[T any] interface {
	Encode(T) ([]byte, error)
	Decode(T) ([]byte, error)
}

type CompressorOpts[T any] struct {
	Engine[T]
}

type Compressor[T any] struct {
	CompressorOpts[T]
	Source T
}

func (c *Compressor[T]) Compress() ([]byte, error) {
	return c.Engine.Encode(c.Source)
}

func (c *Compressor[T]) Decompress() ([]byte, error) {
	return c.Engine.Decode(c.Source)
}
