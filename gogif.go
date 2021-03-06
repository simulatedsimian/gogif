package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"image/gif"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Flags struct {
	Mono bool
	Once bool
	Help bool
}

var config Flags

func init() {
	flag.BoolVar(&config.Help, "h", false, "Display Help")
	flag.BoolVar(&config.Mono, "m", false, "Play in Monochrome mode")
	flag.BoolVar(&config.Once, "o", false, "Play animation once")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: gogif [options] <filename>")
		flag.PrintDefaults()
	}
}

type State struct {
	FrameNumber int
	Config      Flags
	Gif         *gif.GIF
	Attribs     []AttribTable
}

func (s *State) OnInit(gc *GameCore) error {
	mode := termbox.SetOutputMode(termbox.Output256)

	if mode != termbox.Output256 {
		return errors.New("Failed to set output mode")
	}

	cmap := CMapRGB
	if s.Config.Mono {
		cmap = CMapMono
	}
	s.Attribs = mapColours(s.Gif, cmap)

	return nil
}

func (s *State) OnEvent(gc *GameCore, ev *termbox.Event) error {
	if ev.Type == termbox.EventKey {
		if ev.Ch == 'q' {
			gc.DoQuit = true
		}
	}
	return nil
}

func (s *State) OnTick(gc *GameCore) error {
	err := renderFrameHiRes(s.Gif, s.FrameNumber, s.Attribs)
	s.FrameNumber++
	if len(s.Gif.Image) == s.FrameNumber {
		s.FrameNumber = 0
	}
	return err
}

func openFile(name string) (io.ReadCloser, error) {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		resp, err := http.Get(name)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return os.Open(name)
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 || config.Help {
		flag.Usage()
		os.Exit(1)
	}

	f, err := openFile(flag.Args()[0])
	exitOnError(err)

	g, err := gif.DecodeAll(f)
	exitOnError(err)

	state := State{
		Gif:    g,
		Config: config,
	}

	gc := GameCore{
		TickTime: time.Millisecond * 50,
		OnInit:   state.OnInit,
		OnEvent:  state.OnEvent,
		OnTick:   state.OnTick,
	}

	err = gc.Run()
	exitOnError(err)
}
