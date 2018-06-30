package auth

type AccessKey struct {
	AccessKeyId     string
	AccessKeySecret string
	Roles           []string
}
