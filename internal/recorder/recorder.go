// SPDX-FileCopyrightText: 2024 Andrew Pantuso <ajpantuso@gmail.com>
//
// SPDX-License-Identifier: Apache-2.0

package recorder

import (
	"fmt"
	"os"
)

type Recorder interface {
	RecordProduct(Product) error
}

type Product struct {
	Source string
	Name   string
	URL    string
}

func NewDebugRecorder() *DebugRecorder {
	return &DebugRecorder{}
}

type DebugRecorder struct{}

func (r *DebugRecorder) RecordProduct(product Product) error {
	_, err := fmt.Fprintf(os.Stdout, "%+v\n", product)

	return err
}
