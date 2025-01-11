package types

// DescribeImageResponse stores the data of a response
// from an AI image description e.g.
type DescribeImageResponse struct {
	Description string `json:"description" yaml:"description"` // the long description for aria-description maybe
	Label       string `json:"label" yaml:"label"`             // the label for aria-label maybe
}
