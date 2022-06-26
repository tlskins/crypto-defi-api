package api

func BoolPtr(in bool) (out *bool) {
	return &in
}

func StrPtr(in string) (out *string) {
	return &in
}
