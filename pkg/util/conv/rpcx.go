package conv

func ProtoEnumsToRpcxBasePath(enums map[int32]string) (path string) {
	path = ""
	for _, v := range enums {
		if len(v) > 0 {
			path += "/" + v
		}
	}

	return path
}
