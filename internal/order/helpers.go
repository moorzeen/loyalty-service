package order

import (
	"strconv"
)

func ParseOrderNumber(s string) (int64, error) {
	orderNumber, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return orderNumber, err
	}

	n := len(s)
	checksum := 0

	for i := 1; i <= len(s); i++ {
		d, err := strconv.Atoi(string(s[n-i]))
		if err != nil {
			return orderNumber, err
		}

		if i%2 == 0 {
			s := 2 * d
			if s > 9 {
				s -= 9
			}
			checksum += s
		} else {
			checksum += d
		}
	}

	if checksum%10 != 0 {
		return orderNumber, ErrInvalidOrderNumber
	}

	return orderNumber, nil
}
