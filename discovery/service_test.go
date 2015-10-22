package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeDnsEntry(t *testing.T) {

	source := `Stamp\.\.\ \195\132r\ En\ Liten\ Fisk`

	result := decodeDnsEntry(source)

	assert.Equal(t, result, "Stamp.. Är En Liten Fisk")

}
