package utils

const (
	B  = 1
	KB = B * 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

const (
	FileSizeLimit           = 10 * GB
	InitPathMinimalCapacity = 2 * FileSizeLimit
)
