package cos

// BizType defines the COS folder for different business scenarios.
type BizType string

const (
	BizFood     BizType = "food"
	BizFeedback BizType = "feedback"
)

// ValidBizTypes returns all valid business types.
func ValidBizTypes() map[BizType]bool {
	return map[BizType]bool{
		BizFood:     true,
		BizFeedback: true,
	}
}
