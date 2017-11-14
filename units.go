package units

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Prefixes are SI prefixes for encoding and decoding
var Prefixes = []string{"p", "n", "u", "m", "", "K", "M", "G", "T"}

// Orders are the associated orders for each prefix
var Orders = []int{-12, -9, -6, -3, 0, 3, 6, 9, 12}

var prefixMap map[string]int
var orderMap map[int]string

func init() {
	prefixMap := make(map[string]int)
	orderMap := make(map[int]string)
	for i := range Prefixes {
		prefixMap[Prefixes[i]] = Orders[i]
		orderMap[Orders[i]] = Prefixes[i]
	}
}

// MarshalUnit is a helper for common (SI) unit serialisation/marshalling
func MarshalUnit(unit string, value float64) ([]byte, error) {
	// Calculate exponent
	exponent := 0
	if math.Abs(value) > 1 {
		for divided := value; math.Abs(divided) < 10.0; divided = divided / 10.0 {
			exponent++
		}
	} else {
		for multiplied := value; math.Abs(multiplied) > 1.0; multiplied = multiplied * 10.0 {
			exponent--
		}
	}

	prefixExponent := (exponent % 3) * 3
	prefix, ok := orderMap[prefixExponent]
	if !ok {
		return nil, fmt.Errorf("Unsupported prefix for exponent 10^%d", prefixExponent)
	}

	str := fmt.Sprintf("%.2f %s%s", value/math.Pow(10, float64(exponent)), prefix, unit)

	return []byte(str), nil
}

// UnitRegex matches unit strings of the form `[numerator].[denominator] [prefix][unit]` ie. `10.2 dBmV`
var unitRegex = regexp.MustCompile(`^([\-]?[0-9\.]+)[ ]{0,1}([a-zA-Z]+)$`)

// UnmarshalUnit is a helper for common (SI) unit deserialisation/unmarshalling
func UnmarshalUnit(unit string, text []byte) (float64, error) {

	// Match on UnitRegex to check for sane strings
	matches := unitRegex.FindStringSubmatch(string(text))
	if matches == nil {
		return 0.0, fmt.Errorf("Unit must be of the form 'Value PrefixUnit`, ie. '100.2 K%s'", unit)
	}

	// Split value and unit
	valueString := matches[1]
	unitString := matches[2]

	// Check suffix matches
	if !strings.HasSuffix(unitString, unit) {
		return 0.0, fmt.Errorf("Unable to parse unit: '%s' expected suffix: '%s'", unitString, unit)
	}

	// Strip suffix and calculate order from prefix
	prefix := strings.TrimSuffix(unitString, unit)
	order := 0
	if prefix != "" {
		for i := range Prefixes {
			if prefix == Prefixes[i] {
				order = Orders[i]
			}
		}
		if order == 0 {
			return 0.0, fmt.Errorf("Unrecognised SI prefix: '%s' (options: %s)", prefix, strings.Join(Prefixes, ", "))
		}
	}

	// Parse floating point component
	base, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0.0, err
	}

	// Multiply by prefix order
	value := base * math.Pow(10, float64(order))

	return value, nil
}
