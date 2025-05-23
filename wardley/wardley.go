package wardley

import (
	"context"
	"embed"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type BoxModel = dom.BoxModel

type RenderEngine struct {
	ctx     context.Context
	cancel  context.CancelFunc
	closefn func()
}

//go:embed all:dist/*
var dist embed.FS

func CreateListener() (l net.Listener, close func()) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l, func() {
		_ = l.Close()
	}
}

func StartHTTPServer(l net.Listener) {
	var files = fs.FS(dist)
	htmlContent, err := fs.Sub(files, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fs := http.FileServer(http.FS(htmlContent))

	http.Handle("/", fs)
	log.Printf("Listening on :%v\n", l.Addr().(*net.TCPAddr).Port)
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func NewRenderEngine() (*RenderEngine, error) {
	log.Println("Starting http server")

	l, closefn := CreateListener()

	go func() {
		StartHTTPServer(l)
	}()
	log.Println("New Render engine")
	actx, _ := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false), chromedp.WindowSize(1920, 1080))...)
	ctx, cancel := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))

	url := fmt.Sprintf("http://localhost:%v", l.Addr().(*net.TCPAddr).Port)
	log.Printf("Going to %s", url)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url), //"https://onlinewardleymaps.com"),
	)
	log.Println("Returning render engine")
	return &RenderEngine{
		ctx:     ctx,
		cancel:  cancel,
		closefn: closefn,
	}, err
}

func (r *RenderEngine) Render(content string) (string, error) {
	var (
		result string
	)
	log.Println("Running Render")
	var nodes []*cdp.Node
	err := chromedp.Run(r.ctx,
		chromedp.WaitVisible(`document.querySelector("#htmEditor > textarea")`, chromedp.ByJSPath),
		chromedp.Nodes(`document.querySelector("#htmEditor > textarea")`, &nodes, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
	)
	if err != nil {
		return "", err
	}
	log.Println("Moving on")
	err = chromedp.Run(r.ctx,
		chromedp.MouseClickNode(nodes[0]),
		chromedp.KeyEvent("a", chromedp.KeyModifiers(input.ModifierCtrl)),
		chromedp.Sleep(time.Second),
		chromedp.KeyEvent("x", chromedp.KeyModifiers(input.ModifierCtrl)),
		input.InsertText(content),
	)

	if err != nil {
		return "", err
	}
	log.Println("Text inserted")

	err = chromedp.Run(r.ctx,
		chromedp.WaitVisible(`document.querySelector("#top-nav-wrapper > div.MuiBox-root.css-i9gxme > header > div > div.MuiStack-root.css-45az3d > button.MuiButtonBase-root.MuiButton-root.MuiButton-outlined.MuiButton-outlinedError.MuiButton-sizeSmall.MuiButton-outlinedSizeSmall.MuiButton-colorError.MuiButton-root.MuiButton-outlined.MuiButton-outlinedError.MuiButton-sizeSmall.MuiButton-outlinedSizeSmall.MuiButton-colorError.css-id02k7")`, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
	)
	if err != nil {
		return "", err
	}

	log.Println("Wait successful")
	err = chromedp.Run(r.ctx,
		chromedp.OuterHTML(`document.querySelector("#map > div.wardley.jss1 > svg")`, &result, chromedp.ByJSPath),
	)
	if err != nil {
		return "", err
	}

	log.Println("Node Captured")

	return strings.ReplaceAll(result, "&nbsp;", "&#160;"), err
}

func (r *RenderEngine) RenderAsScaledPng(content string, scale float64) ([]byte, *BoxModel, error) {
	var (
		result_in_bytes []byte
		model           *dom.BoxModel
	)

	_, err := r.Render(content)
	if err != nil {
		return result_in_bytes, interface{}(model).(*BoxModel), err
	}
	log.Println("Running scaled PNG")
	err = chromedp.Run(r.ctx,
		chromedp.ScreenshotScale(`document.querySelector("#map > div.wardley.jss1 > svg")`, scale, &result_in_bytes, chromedp.ByJSPath),
		chromedp.Dimensions(`document.querySelector("#map > div.wardley.jss1 > svg")`, &model, chromedp.ByJSPath),
	)
	return result_in_bytes, interface{}(model).(*BoxModel), err
}

func (r *RenderEngine) RenderAsPng(content string) ([]byte, *BoxModel, error) {
	return r.RenderAsScaledPng(content, 1.0)
}

func (r *RenderEngine) Cancel() {
	r.cancel()
}
