package numbers

func MinInt(nums ...int) int {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MinInt32(nums ...int32) int32 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MinInt64(nums ...int64) int64 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MinUint(nums ...uint) uint {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MinUint32(nums ...uint32) uint32 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MinUint64(nums ...uint64) uint64 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] < n {
			n = nums[i]
		}
	}
	return n
}

func MaxInt(nums ...int) int {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func MaxInt32(nums ...int32) int32 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func MaxInt64(nums ...int64) int64 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func MaxUint(nums ...uint) uint {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func MaxUint32(nums ...uint32) uint32 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func MaxUint64(nums ...uint64) uint64 {
	n := nums[0]
	for i, l := 1, len(nums); i < l; i++ {
		if nums[i] > n {
			n = nums[i]
		}
	}
	return n
}

func LimitInt(n, min, max int) int {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func LimitInt32(n, min, max int32) int32 {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func LimitInt64(n, min, max int64) int64 {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func LimitUint(n, min, max uint) uint {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func LimitUint32(n, min, max uint32) uint32 {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func LimitUint64(n, min, max uint64) uint64 {
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}
