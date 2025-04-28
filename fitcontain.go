package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var RED, GREEN, BLUE, YELLOW = color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 255, 255}

// use the poker case as the basis of a custom diagnostic widget that has adornments that react to changes in the size and location of the image when the canvas is resized
func FitContain(media embed.FS, screenSize fyne.Size, _ float32) *fyne.Container {
	return container.NewStack(newDiagnosticWidget(loadImage(media, screenSize)))
}

type crossHair struct {
	horiz, vert *canvas.Line
}

type diagBox struct {
	box [5]*canvas.Rectangle
}

type diagData struct {
	text []*canvas.Text
}

// use cross hairs over an image that is resized and repositioned using FillContain and is responsive to window size changes
type diagnosticWidget struct {
	widget.BaseWidget
	// the image defines the size of the widget, it is used as the basis of the diagnostic
	img *canvas.Image
	// the image will grow and shrink relative to it's initial size
	imgInitialSize fyne.Size
	// middle of the image
	crossHair
	// corners of image
	diagBox
	// diagnostic data
	diagData
	// grid for determining size of image
	grid [20]*canvas.Line
	// circles to position
	circle [2]*canvas.Circle
}

// create a diagnostic widget with location markers
func newDiagnosticWidget(img *canvas.Image) *diagnosticWidget {
	img.ScaleMode = canvas.ImageScaleFastest
	img.FillMode = canvas.ImageFillContain
	t := &diagnosticWidget{
		BaseWidget:     widget.BaseWidget{},
		img:            img,
		imgInitialSize: img.Size(),
		crossHair: crossHair{
			horiz: canvas.NewLine(RED),
			vert:  canvas.NewLine(RED),
		},
		diagBox: diagBox{
			box: [5]*canvas.Rectangle{
				canvas.NewRectangle(RED),
				canvas.NewRectangle(RED),
				canvas.NewRectangle(RED),
				canvas.NewRectangle(RED),
				canvas.NewRectangle(RED),
			},
		},
		diagData: diagData{
			text: []*canvas.Text{canvas.NewText("", GREEN), canvas.NewText("", GREEN)},
		},
		grid:   [20]*canvas.Line{},
		circle: [2]*canvas.Circle{canvas.NewCircle(color.Transparent), canvas.NewCircle(color.Transparent)},
	}
	for _, box := range t.diagBox.box {
		box.FillColor = color.Transparent
		box.StrokeColor = RED
		box.StrokeWidth = 1

	}
	for _, text := range t.diagData.text {
		text.Alignment = fyne.TextAlignLeading
		text.TextSize = 14
	}
	for i := range len(t.grid) {
		line := canvas.NewLine(color.Transparent)
		line.StrokeColor = BLUE
		line.StrokeWidth = 1
		t.grid[i] = line
	}
	for i := range len(t.circle) {
		if i == 0 {
			t.circle[i].StrokeColor = RED
		} else {
			t.circle[i].StrokeColor = YELLOW
		}
		t.circle[i].StrokeWidth = 1
	}

	t.ExtendBaseWidget(t)
	return t
}

type diagLoc int

const (
	LOC_MID diagLoc = iota
	LOC_UL
	LOC_UR
	LOC_LL
	LOC_LR
)

type diagnosticWidgetRenderer struct{ diagnosticWidget *diagnosticWidget }

func (t *diagnosticWidget) Hide() { t.BaseWidget.Hide() }

func (t *diagnosticWidget) Show() { t.BaseWidget.Show() }

func (t *diagnosticWidget) MinSize() fyne.Size { return fyne.NewSquareSize(100) }

func (t *diagnosticWidget) Move(pos fyne.Position) { t.BaseWidget.Move(pos) }

func (t *diagnosticWidget) Refresh() { t.BaseWidget.Refresh() }

func (t *diagnosticWidget) Resize(size fyne.Size) { t.BaseWidget.Resize(size) }

func (t *diagnosticWidget) CreateRenderer() fyne.WidgetRenderer {
	return &diagnosticWidgetRenderer{diagnosticWidget: t}
}

func (r *diagnosticWidgetRenderer) Layout(size fyne.Size) {
	r.diagnosticWidget.img.Resize(size)
	r.updateDiagMarkers(size)
	r.updateCornerMarkers(size)
	r.updateDiagData(size)
}

func (r *diagnosticWidgetRenderer) updateDiagData(winSize fyne.Size) {
	text := r.diagnosticWidget.diagData.text
	text[0].Text = fmt.Sprintf("winSize: %v, iis: %v", winSize, r.diagnosticWidget.imgInitialSize)
	text[0].Move(fyne.NewPos(0, 0))
	text[0].Refresh()
}

func (r *diagnosticWidgetRenderer) updateDiagMarkers(size fyne.Size) {
	r.drawCrossHairsAndDiagnosticObjects(LOC_MID, size)
}

func drawObject(o fyne.CanvasObject, pt, sz fyne.Vector2) {
	x, y := pt.Components()
	w, h := sz.Components()
	o.Move(fyne.NewPos(x, y))
	o.Resize(fyne.NewSize(w, h))
}

func XLatePosByVector(p fyne.Position, v fyne.Vector2) fyne.Position {
	w, h := v.Components()
	return fyne.NewPos(p.X*w, p.Y*h)
}

func ScalePosByScalar(p fyne.Vector2, factor float32) fyne.Position {
	x, y := p.Components()
	return fyne.NewPos(x*factor, y*factor)
}

func XLateVectorToPos(p1, p2 fyne.Vector2) fyne.Position {
	x1, y1 := p1.Components()
	x2, y2 := p2.Components()
	return fyne.NewPos(x1+x2, y1+y2)
}

func ScaleSize(s fyne.Size, factor float32) fyne.Size {
	return fyne.NewSize(s.Width*factor, s.Height*factor)
}

func FitAndCenterPoint(win, box fyne.Size) fyne.Position {
	return fyne.NewPos((win.Width-box.Width)/2, (win.Height-box.Height)/2)
}

// this draws cross hairs on the window, a box at the center of the window, boxes at the corners of the window and a circle in the middle of the UL quadrant of the window
// a circle representing an object in the image is drawn, scaled to match the new size of the image and placed in the window where it exists on the image
// the circle is where a button would be placed and sized
func (r *diagnosticWidgetRenderer) drawCrossHairsAndDiagnosticObjects(loc diagLoc, size fyne.Size) {
	if loc == LOC_MID {
		// cross hairs
		drawObject(r.diagnosticWidget.vert, fyne.NewPos(size.Width/2, 0), fyne.NewSize(0, size.Height))
		drawObject(r.diagnosticWidget.horiz, fyne.NewPos(0, size.Height/2), fyne.NewSize(size.Width, 0))

		// center box in UL quadrant
		//   calculate the quadrant and 1/2 of the quadrant sizes
		sz, sz2 := fyne.NewSquareSize(20), fyne.NewSquareSize(10)
		//   translate the point to the center of the area
		pt := fyne.NewPos(size.Width/4, size.Height/4).Subtract(sz2)
		drawObject(r.diagnosticWidget.circle[0], pt, sz)

		// this circle should stay in the middle of the UL quadrant of the image and have the correct size
		p := fyne.NewSquareOffsetPos(0.25)

		// the initial size of the image
		iis := r.diagnosticWidget.imgInitialSize

		// the position of the point in the initial size of the image
		pAbs := XLatePosByVector(p, iis)

		// calculate the scale constrained by the image aspect ratio that results from resizing the window
		S := min(size.Width/iis.Width, size.Height/iis.Height)

		// calculate the size of the image that fits in the window
		imgSizeCur := ScaleSize(iis, S)

		// calculate the of center the image that fits in the window
		curImgCenter := FitAndCenterPoint(size, imgSizeCur)

		// translate the position of the point in the initial size of the image to the scaled location in the window
		pRel := XLateVectorToPos(curImgCenter, ScalePosByScalar(pAbs, S))

		// resize the point (i.e. circle) to match the size of the image that fits in the window and center the circle
		csz, csz2 := ScaleSize(sz, S), ScaleSize(sz, S/2)
		pt = fyne.NewPos(pRel.X, pRel.Y).Subtract(csz2)
		drawObject(r.diagnosticWidget.circle[1], pt, csz)
	}
}

func (r *diagnosticWidgetRenderer) updateCornerMarkers(widgetSize fyne.Size) {
	d := fyne.NewSize(widgetSize.Width-r.diagnosticWidget.imgInitialSize.Width, widgetSize.Height-r.diagnosticWidget.imgInitialSize.Height)
	dp := fyne.NewSize(widgetSize.Width/r.diagnosticWidget.imgInitialSize.Width, widgetSize.Height/r.diagnosticWidget.imgInitialSize.Height)

	// the image is landscape so the new image size to fit in the window is always scaled by change in window width (dp = delta percent)
	nImgSize := fyne.NewSize(r.diagnosticWidget.imgInitialSize.Width*dp.Width, r.diagnosticWidget.imgInitialSize.Height*dp.Width)
	// d.Height/2 is vertical padding when the window is growing vertically
	// pad := fyne.NewSize(0, d.Height/2)
	pad := fyne.NewSize(0, 0)
	text := r.diagnosticWidget.diagData.text
	text[1].Text = fmt.Sprintf("dw,dh: %v, dw%%,dh%%: %v, nImgSize: %v",
		d,
		dp,
		nImgSize,
	)
	text[1].Move(fyne.NewPos(0, text[0].Position().Y+text[0].TextSize*1.1))
	text[1].Refresh()

	// position of the box for each loc
	var pt fyne.Position
	for i, loc := range []diagLoc{LOC_UL, LOC_UR, LOC_LL, LOC_LR, LOC_MID} {
		box := r.diagnosticWidget.diagBox.box[i]
		sz := fyne.NewSquareSize(20)
		half_sz := fyne.NewSquareSize(20 / 2)
		switch loc {
		case LOC_UL:
			pt = fyne.NewPos(0, 0).Subtract(pad)
		case LOC_UR:
			pt = fyne.NewPos(widgetSize.Width, 0).SubtractXY(sz.Width, 0).Subtract(pad)
		case LOC_LL:
			pt = fyne.NewPos(0, widgetSize.Height).SubtractXY(0, sz.Height).Add(pad)
		case LOC_LR:
			pt = fyne.NewPos(widgetSize.Width, widgetSize.Height).SubtractXY(sz.Width, sz.Height).Add(pad)
		case LOC_MID:
			pt = fyne.NewPos((widgetSize.Width)/2, (widgetSize.Height)/2).Subtract(half_sz)
		default:
			return
		}
		box.Move(pt)
		box.Resize(sz)
	}
}

func (r *diagnosticWidgetRenderer) MinSize() fyne.Size { return r.diagnosticWidget.MinSize() }

func (r *diagnosticWidgetRenderer) Refresh() { r.Layout(r.diagnosticWidget.Size()) }

func (r *diagnosticWidgetRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{
		r.diagnosticWidget.img,
		r.diagnosticWidget.horiz,
		r.diagnosticWidget.vert,
	}
	for _, text := range r.diagnosticWidget.diagData.text {
		objects = append(objects, text)
	}
	for _, rect := range r.diagnosticWidget.diagBox.box {
		objects = append(objects, rect)
	}
	for _, line := range r.diagnosticWidget.grid {
		objects = append(objects, line)
	}
	for _, circle := range r.diagnosticWidget.circle {
		objects = append(objects, circle)
	}
	return objects
}

func (r *diagnosticWidgetRenderer) Destroy() {}

func loadImage(media embed.FS, size fyne.Size) (canvasImg *canvas.Image) {
	fileName := "media/fitcontain-diag.png"
	data, err := media.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	canvasImg = canvas.NewImageFromImage(img)
	canvasImg.FillMode, canvasImg.ScaleMode = canvas.ImageFillContain, canvas.ImageScaleFastest
	canvasImg.Resize(size)
	return canvasImg
}
