package main

import (
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"image/color"
	"image/gif"
	"os"
	"time"
)

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type AttribVals struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

type AttribTable [256]AttribVals

type ColourMapFunc func(color.Color) AttribVals

func test(c color.Color) AttribVals {

	rgba := color.RGBAModel.Convert(c).(color.RGBA)

	return AttribVals{fg: termbox.Attribute(rgba.R), bg: termbox.Attribute(rgba.B)}
}

func CMapMono(c color.Color) AttribVals {

	g := color.GrayModel.Convert(c).(color.Gray)

	return AttribVals{
		fg: termbox.Attribute(g.Y/11 + 1),
		bg: termbox.Attribute(g.Y/11 + 1),
	}
}

func mapColours(g *gif.GIF, cmap ColourMapFunc) []AttribTable {
	var attribs []AttribTable

	for f := 0; f < len(g.Image); f++ {
		var at AttribTable
		for i := 0; i < len(g.Image[f].Palette); i++ {
			at[i] = cmap(g.Image[f].Palette[i])
		}
		attribs = append(attribs, at)
	}

	return attribs
}

func renderFrame(g *gif.GIF, framenum int, attribs []AttribTable) error {

	width, height := termbox.Size()

	if width > g.Image[framenum].Rect.Dx() {
		width = g.Image[framenum].Rect.Dx()
	}

	if height > g.Image[framenum].Rect.Dy() {
		height = g.Image[framenum].Rect.Dy()
	}

	for y := 0; y < height; y++ {
		lineOffset := g.Image[framenum].Stride * y
		for x := 0; x < width; x++ {
			attr := attribs[framenum][g.Image[framenum].Pix[x+lineOffset]]
			termbox.SetCell(x, y, ' ', attr.fg, attr.bg)
		}
	}

	return nil
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "Usage: gogif <filename>")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	exitOnError(err)

	g, err := gif.DecodeAll(f)
	exitOnError(err)

	gc := GameCore{}
	gc.TickTime = time.Millisecond * 50

	gc.OnInit = func(gc *GameCore) error {
		mode := termbox.SetOutputMode(termbox.OutputGrayscale)

		if mode != termbox.OutputGrayscale {
			return errors.New("termbox.OutputGrayscale")
		}

		return nil
	}

	gc.OnEvent = func(gc *GameCore, ev *termbox.Event) error {
		if ev.Type == termbox.EventKey {
			if ev.Ch == 'q' {
				gc.DoQuit = true
			}
		}
		return nil
	}

	frameNumber := 0

	attribs := mapColours(g, CMapMono)

	gc.OnTick = func(gc *GameCore) error {
		err := renderFrame(g, frameNumber, attribs)
		frameNumber++
		if len(g.Image) == frameNumber {
			frameNumber = 0
			//gc.DoQuit = true
		}
		return err
	}

	gc.Run()
}
