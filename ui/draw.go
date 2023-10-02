package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type ScaleType string

const (
	ScaleTypeStretch         ScaleType = "stretch"
	ScaleTypeKeepAspectRatio ScaleType = "keepaspect"
	ScaleTypeKeepSize        ScaleType = "keepsize"
)

func (s ScaleType) apply(geom *ebiten.GeoM, destSize, sourceSize, maxSize image.Point) image.Point {
	if s == ScaleTypeKeepSize {
		return sourceSize
	}

	if maxSize.X == 0 {
		maxSize.X = destSize.X
	}
	if maxSize.Y == 0 {
		maxSize.Y = destSize.Y
	}

	targetSize := destSize

	if s == ScaleTypeKeepAspectRatio {
		targetSize = image.Point{
			X: maxSize.X,
			Y: int(float64(maxSize.X) * (float64(sourceSize.Y) / float64(sourceSize.X))),
		}

		if targetSize.Y > maxSize.Y {
			targetSize = image.Point{
				X: int(float64(maxSize.Y) * (float64(sourceSize.X) / float64(sourceSize.Y))),
				Y: maxSize.Y,
			}
		}
	}

	geom.Scale(
		float64(targetSize.X)/float64(sourceSize.X),
		float64(targetSize.Y)/float64(sourceSize.Y))

	return targetSize
}

type HorizontalAlign string

const (
	HorizontalAlignLeft   HorizontalAlign = "left"
	HorizontalAlignRight  HorizontalAlign = "right"
	HorizontalAlignCenter HorizontalAlign = "center"
)

func (h HorizontalAlign) apply(geom *ebiten.GeoM, destSize, targetSize image.Point) {
	switch h {
	case HorizontalAlignRight:
		geom.Translate(float64(destSize.X-targetSize.X), 0)
	case HorizontalAlignCenter:
		geom.Translate(float64((destSize.X-targetSize.X)/2), 0)
	case HorizontalAlignLeft:
		fallthrough
	default:
		break
	}
}

type VerticalAlign string

const (
	VerticalAlignTop    VerticalAlign = "top"
	VerticalAlignBottom VerticalAlign = "bottom"
	VerticalAlignCenter VerticalAlign = "center"
)

func (v VerticalAlign) apply(geom *ebiten.GeoM, destSize, targetSize image.Point) {
	switch v {
	case VerticalAlignBottom:
		geom.Translate(0, float64(destSize.Y-targetSize.Y))
	case VerticalAlignCenter:
		geom.Translate(0, float64((destSize.Y-targetSize.Y)/2))
	case VerticalAlignTop:
		fallthrough
	default:
		break
	}
}

type DrawImageOptions struct {
	ScaleType              ScaleType
	MaxSize                image.Point
	HorizontalAlign        HorizontalAlign
	VerticalAlign          VerticalAlign
	Filter                 ebiten.Filter
	TargetRect             image.Rectangle
	ColorScale             colorm.ColorM
	ProportionalTranslateY float64
}

func DrawImage(dest *ebiten.Image, source *ebiten.Image, options DrawImageOptions) image.Point {
	destSize := dest.Bounds().Size()
	sourceSize := source.Bounds().Size()
	geom := ebiten.GeoM{}

	if !options.TargetRect.Empty() {
		destSize = options.TargetRect.Bounds().Size()
		geom.Translate(float64(options.TargetRect.Min.X), float64(options.TargetRect.Min.Y))
	}

	targetSize := options.ScaleType.apply(&geom, destSize, sourceSize, options.MaxSize)
	options.HorizontalAlign.apply(&geom, destSize, targetSize)
	options.VerticalAlign.apply(&geom, destSize, targetSize)
	geom.Translate(0, float64(targetSize.Y)*options.ProportionalTranslateY)

	colorm.DrawImage(dest, source, options.ColorScale, &colorm.DrawImageOptions{
		GeoM:   geom,
		Filter: options.Filter,
	})

	return targetSize
}
