package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtoBuilder(t *testing.T) {
	var pb ProtoBuilder

	_ = pb.Varint(1, 128)
	_ = pb.Varint(2, 1234567890)
	_ = pb.Varint(3, 1234567890123456789)
	_ = pb.String(4, "Hello")
	_ = pb.Bytes(5, []byte{1, 2, 3})
	assert.Equal(t, "CIABENKF2MwEGJWCpu_HnoSRESIFSGVsbG8qAwECAw%3D%3D", pb.ToURLEncodedBase64())
}
