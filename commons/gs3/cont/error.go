package cont

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gs3"
)

type HashAlreadyExistsError struct {
	Object *gs3.ObjectInfo
}

func (e *HashAlreadyExistsError) Error() string {
	return fmt.Sprintf("hash already exists: %s", e.Object.Key)
}
