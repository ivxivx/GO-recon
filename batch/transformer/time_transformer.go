package transformer

import (
	"fmt"
	"time"
)

type TimeTransformer struct {
	InputFormat  string
	OutputFormat string
}

var _ FieldTransformer = (*TimeTransformer)(nil)

func (tf *TimeTransformer) Transform(value string) (string, error) {
	parsed, err := time.Parse(tf.InputFormat, value)
	if err != nil {
		return "", fmt.Errorf("time transformer error: %w", err)
	}

	transformed := parsed.Format(tf.OutputFormat)

	return transformed, nil
}
