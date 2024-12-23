package cont

import (
	"fmt"
	"github.com/qiafan666/gotato/s3"
)

type HashAlreadyExistsError struct {
	Object *s3.ObjectInfo
}

func (e *HashAlreadyExistsError) Error() string {
	return fmt.Sprintf("hash already exists: %s", e.Object.Key)
}
