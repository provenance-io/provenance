package export

import "fmt"

// MaxEventTagLength is the maximum length that an event tag can have.
// 100 was chosen because that's what we used for the external ids.
const MaxEventTagLength = 100

// ValidateEventTag makes sure an event tag is okay.
func ValidateEventTag(eventTag string) error {
	if len(eventTag) > MaxEventTagLength {
		return fmt.Errorf("invalid event tag %q (length %d): exceeds max length %d",
			eventTag[:5]+"..."+eventTag[len(eventTag)-5:], len(eventTag), MaxEventTagLength)
	}
	return nil
}
