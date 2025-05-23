package wardley

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

type BoxModel = dom.BoxModel

type RenderEngine struct {
	ctx    context.Context
	cancel context.CancelFunc
}

//go:embed all:dist/*
var dist embed.FS

// StartHTTPServer starts a HTTP webserver that hosts the onlinewardleymaps.com react app locally
func StartHTTPServer(l net.Listener) {
	var files = fs.FS(dist)
	htmlContent, err := fs.Sub(files, "dist")
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
	fs := http.FileServer(http.FS(htmlContent))

	http.Handle("/", fs)
	log.Info().Int("port", l.Addr().(*net.TCPAddr).Port).Msg("Server started Listening")
	err = http.Serve(l, nil)
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

// NewRenderEngine returns a RenderEngine which can be use to render Wardley Maps
func NewRenderEngine(port string) (*RenderEngine, error) {
	log.Debug().Msg("New Render engine")

	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}

	go func() {
		StartHTTPServer(l)
	}()

	log.Debug().Msg("HTTP server started")

	actx, _ := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", true), chromedp.WindowSize(1920, 1080))...)
	ctx, cancel := chromedp.NewContext(actx,
		chromedp.WithLogf(log.Info().Msgf),
		chromedp.WithErrorf(log.Error().Msgf),
		chromedp.WithDebugf(log.Trace().Msgf),
	)
	url := fmt.Sprintf("http://localhost:%v", l.Addr().(*net.TCPAddr).Port)

	log.Debug().Str("url", url).Msg("Navigating to owm webapp")

	err = chromedp.Run(ctx,
		chromedp.Sleep(time.Second),
		chromedp.Navigate(url),
	)
	return &RenderEngine{
		ctx:    ctx,
		cancel: cancel,
	}, err
}

// Render returns a rendered SVG for a given Wardley Map
func (r *RenderEngine) Render(content string) ([]byte, error) {
	var (
		result     string
		nodes      []*cdp.Node
		errorNodes []*cdp.Node
	)

	log.Debug().Msg("Running Render")

	if content == "" {
		return nil, fmt.Errorf("no map provided")
	}

	err := chromedp.Run(r.ctx,
		chromedp.WaitVisible(`document.querySelector("#htmEditor > textarea")`, chromedp.ByJSPath),
		chromedp.Nodes(`document.querySelector("#htmEditor > textarea")`, &nodes, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
	)
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("Selected textarea")

	// Paste Wardley Map into textarea
	err = chromedp.Run(r.ctx,
		chromedp.Click(`//*[@id="new-menu-button"]`),
		chromedp.MouseClickNode(nodes[0]),
		chromedp.Sleep(time.Second),
		input.InsertText(content),
	)

	if err != nil {
		return nil, err
	}
	log.Debug().Str("content", content).Msg("Text inserted")

	// Check if syntax errors exist
	err = chromedp.Run(r.ctx,
		chromedp.WaitVisible(`document.querySelector("#top-nav-wrapper > div.MuiBox-root.css-i9gxme > header > div > div.MuiStack-root.css-45az3d > button.MuiButtonBase-root.MuiButton-root.MuiButton-outlined.MuiButton-outlinedError.MuiButton-sizeSmall.MuiButton-outlinedSizeSmall.MuiButton-colorError.MuiButton-root.MuiButton-outlined.MuiButton-outlinedError.MuiButton-sizeSmall.MuiButton-outlinedSizeSmall.MuiButton-colorError.css-id02k7")`, chromedp.ByJSPath),
		chromedp.Nodes(".ace_error", &errorNodes, chromedp.ByQuery, chromedp.AtLeast(0)),
	)
	if err != nil {
		return nil, err
	}

	if len(errorNodes) > 0 {
		return nil, fmt.Errorf("syntax error")
	}

	log.Info().Msg("Syntax checked and correct")

	err = chromedp.Run(r.ctx,
		chromedp.OuterHTML(`document.querySelector("#map > div.wardley.jss1 > svg")`, &result, chromedp.ByJSPath),
	)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("Node Captured")

	return []byte(strings.ReplaceAll(result, "&nbsp;", "&#160;")), nil
}

// RenderAsScaledPNG returns a PNG with a defined scaling factor from a given Wardley Map
func (r *RenderEngine) RenderAsScaledPng(content string, scale float64) ([]byte, *BoxModel, error) {
	var (
		resultInBytes []byte
		model         *dom.BoxModel
	)

	_, err := r.Render(content)
	if err != nil {
		return nil, nil, err
	}
	log.Debug().Float64("scale", scale).Msg("Screenshotting PNG")
	err = chromedp.Run(r.ctx,
		chromedp.ScreenshotScale(`document.querySelector("#map > div.wardley.jss1 > svg")`, scale, &resultInBytes, chromedp.ByJSPath),
		chromedp.Dimensions(`document.querySelector("#map > div.wardley.jss1 > svg")`, &model, chromedp.ByJSPath),
	)
	if err != nil {
		return nil, nil, err
	}
	return resultInBytes, model, nil
}

// RenderAsPNG returns a PNG with a fixed scaling factor of 1.0 from a given Wardley Map
func (r *RenderEngine) RenderAsPng(content string) ([]byte, *BoxModel, error) {
	return r.RenderAsScaledPng(content, 1.0)
}

func (r *RenderEngine) Cancel() {
	r.cancel()
}
