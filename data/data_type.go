package data

type Ret struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RetDescribeFile struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	FileStream []byte `json:"file_stream"`
}

type RetCreateFile struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	FileMountPath string `json:"file_mount_path"`
}
