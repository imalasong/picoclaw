//go:build !cgo || !matrix_crypto

package matrix

import (
	"context"
	"fmt"
)

func (c *MatrixChannel) maybeInitCrypto(_ context.Context) error {
	return fmt.Errorf("matrix crypto is disabled (build with tags: matrix_crypto and enable cgo)")
}
