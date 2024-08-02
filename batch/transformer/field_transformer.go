package transformer

type FieldTransformer interface {
	Transform(value string) (string, error)
}
