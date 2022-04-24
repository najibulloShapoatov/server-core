package uuid

const (
	variantIndex = 8
	versionIndex = 6
)

// Array is a clean UUID type for simpler UUID versions
type Array [length]byte

// Size returns the length of the array
func (Array) Size() int {
	return length
}

// Version returns the version of the uuid array
func (o Array) Version() int {
	return int(o[versionIndex]) >> 4
}

func (o *Array) setVersion(pVersion int) {
	o[versionIndex] &= 0x0F
	o[versionIndex] |= byte(pVersion) << 4
}

// Variant returns the variant of the uuid array
func (o *Array) Variant() byte {
	return variant(o[variantIndex])
}

func (o *Array) setVariant(pVariant byte) {
	setVariant(&o[variantIndex], pVariant)
}

// Unmarshal decodes the given byte slice into a uuid array
func (o *Array) Unmarshal(pData []byte) {
	copy(o[:], pData)
}

// Bytes returns the uuid array as a byte slice
func (o *Array) Bytes() []byte {
	return o[:]
}

// String prints the uuid array formatted with the standard format
func (o Array) String() string {
	return formatter(&o, format)
}

// Format prints the uuid array formatted with the given format
func (o Array) Format(pFormat string) string {
	return formatter(&o, pFormat)
}

// Set the three most significant bits (bits 0, 1 and 2) of the
// sequenceHiAndVariant equivalent in the array to reservedRFC4122.
func (o *Array) setRFC4122Variant() {
	o[variantIndex] &= 0x3F
	o[variantIndex] |= reservedRFC4122
}

// MarshalBinary marshals the UUID bytes into a slice
func (o *Array) MarshalBinary() ([]byte, error) {
	return o.Bytes(), nil
}

// UnmarshalBinary un-marshals the data bytes into the UUID.
func (o *Array) UnmarshalBinary(pData []byte) error {
	return UnmarshalBinary(o, pData)
}
