package kafka

import (
	"github.com/kxddry/lectura/shared/entities/summarized"
	"github.com/kxddry/lectura/shared/entities/transcribed"
	"github.com/kxddry/lectura/shared/entities/uploaded"
)

type Pipeline[R_, W_ uploaded.Record | transcribed.Record | summarized.Record] struct {
	R Reader[R_]
	W Writer[W_]
}

func NewPipeline[R, W uploaded.Record | transcribed.Record | summarized.Record](r Reader[R], w Writer[W]) Pipeline[R, W] {
	return Pipeline[R, W]{
		R: r,
		W: w,
	}
}
