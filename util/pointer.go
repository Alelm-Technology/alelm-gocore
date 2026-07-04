package util

func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func Float64Ptr(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func IntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}
