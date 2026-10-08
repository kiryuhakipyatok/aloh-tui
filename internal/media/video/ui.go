package video

import (
	"image"
	"sync"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type ui struct {
	frame    atomic.Value
	img      image.Image
	window   *app.Window
	windowed atomic.Bool

	termFrameCount atomic.Int32

	resizedFrame *image.NRGBA

	mu sync.RWMutex
}

func (ui *ui) proccessWindow() {
	var ops op.Ops

	for {
		w := ui.window
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			ui.mu.Lock()
			if w == ui.window {
				ui.windowed.Store(false)
				ui.window = nil
			}
			ui.mu.Unlock()
			return
		case app.FrameEvent:

			gtx := app.NewContext(&ops, e)

			ui.mu.RLock()
			img := ui.img
			ui.mu.RUnlock()
			if img != nil {

				is := img.Bounds().Size()

				if is.X > 0 && is.Y > 0 {
					scale := e.Size

					diffW := float32(scale.X) / float32(is.X)

					diffY := float32(scale.Y) / float32(is.Y)

					op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(diffW, diffY))).Add(&ops)

					i := paint.NewImageOp(img)
					i.Add(&ops)
					paint.PaintOp{}.Add(gtx.Ops)
				}

			}
			e.Frame(gtx.Ops)
			img = nil
			ops.Reset()
		}

	}
}
