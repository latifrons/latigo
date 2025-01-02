package decimal_tools

import "github.com/shopspring/decimal"

func ValueOrZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

func NilOrCloneValuePointer(value *decimal.Decimal) *decimal.Decimal {
	if value == nil {
		return nil
	}
	if value.IsZero() {
		return nil
	}
	v := *value
	return &v
}
func ZeroToNilOrClone(value decimal.Decimal) *decimal.Decimal {
	if value.IsZero() {
		return nil
	}
	v := value
	return &v
}

func SafeAdd(v1 *decimal.Decimal, v2 *decimal.Decimal) *decimal.Decimal {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	c := v1.Add(*v2)
	return &c
}

func IsZeroOrNil(value *decimal.Decimal) bool {
	if value == nil {
		return true
	}
	return value.IsZero()
}
