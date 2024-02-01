package truenas

func flattenInt64List(list []int64) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, num := range list {
		result = append(result, num)
	}
	return result
}

func flattenInt32List(list []int32) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, num := range list {
		result = append(result, num)
	}
	return result
}

func flattenStringList(list []string) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, s := range list {
		result = append(result, s)
	}
	return result
}

func getStringPtr(s string) *string {
	val := s
	return &val
}

func getInt64Ptr(i int64) *int64 {
	val := i
	return &val
}

func getInt32Ptr(i int32) *int32 {
	val := i
	return &val
}

func getBoolPtr(b bool) *bool {
	val := b
	return &val
}

func expandStrings(items []interface{}) []string {
	result := make([]string, 0, len(items))

	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}

func convertStringMap(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	for k, val := range v {
		m[k] = val.(string)
	}
	return m
}

func expandIntegers(items []interface{}) []int32 {
	result := make([]int32, 0, len(items))

	for _, item := range items {
		result = append(result, int32(item.(int)))
	}
	return result
}
