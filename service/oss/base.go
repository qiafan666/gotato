package oss

type MemoryParameter struct {
	Storage                     int64 `json:"storage"`                        // 获取Bucket的总存储量，单位为字节。
	ObjectCount                 int64 `json:"object_count"`                   // 获取Bucket中总的Object数量。
	MultipartUploadCount        int64 `json:"multipart_upload_count"`         // 获取Bucket中已经初始化但还未完成（Complete）或者还未中止（Abort）的Multipart Upload数量。
	LiveChannelCount            int64 `json:"live_channel_count"`             // 获取Bucket中Live Channel的数量。
	LastModifiedTime            int64 `json:"last_modified_time"`             // 此次调用获取到的存储信息的时间点，格式为时间戳，单位为秒。
	StandardStorage             int64 `json:"standard_storage"`               // 获取标准存储类型Object的存储量，单位为字节。
	StandardObjectCount         int64 `json:"standard_object_count"`          // 获取标准存储类型的Object的数量。
	InfrequentAccessStorage     int64 `json:"infrequent_access_storage"`      // 获取低频存储类型Object的计费存储量，单位为字节。
	InfrequentAccessRealStorage int64 `json:"infrequent_access_real_storage"` // 获取低频存储类型Object的实际存储量，单位为字节。
	InfrequentAccessObjectCount int64 `json:"infrequent_access_object_count"` // 获取低频存储类型的Object数量。
	ArchiveStorage              int64 `json:"archive_storage"`                // 获取归档存储类型Object的计费存储量，单位为字节。
	ArchiveRealStorage          int64 `json:"archive_real_storage"`           // 获取归档存储类型Object的实际存储量，单位为字节。
	ArchiveObjectCount          int64 `json:"archive_object_count"`           // 获取归档存储类型的Object数量。
	ColdArchiveStorage          int64 `json:"cold_archive_storage"`           // 获取冷归档存储类型Object的计费存储量，单位为字节。
	ColdArchiveRealStorage      int64 `json:"cold_archive_real_storage"`      // 获取冷归档存储类型Object的实际存储量，单位为字节。
	ColdArchiveObjectCount      int64 `json:"cold_archive_object_count"`      // 获取冷归档存储类型的Object数量。
}
