package cont

const (
	// HashPath 默认为 "gotato/data/hash/"
	hashPath = "gotato/data/hash/"

	// TempPath 默认为 "gotato/temp/"
	tempPath = "gotato/temp/"

	// DirectPath 默认为 "gotato/direct"
	DirectPath = "gotato/direct"

	// UploadTypeMultipart represents the identifier for multipart uploads,
	// allowing large files to be uploaded in chunks.
	UploadTypeMultipart = 1

	// UploadTypePresigned signifies the use of presigned URLs for uploads,
	// facilitating secure, authorized file transfers without requiring direct access to the storage credentials.
	UploadTypePresigned = 2

	// PartSeparator is used as a delimiter in multipart upload processes,
	// separating individual file parts.
	partSeparator = ","
)
