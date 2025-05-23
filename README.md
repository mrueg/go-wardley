# go-wardley 

[go-wardley][] is a library that creates Wardley Maps using the syntax from [OnlineWardleyMaps][].
It uses [chromedp][] to execute [OnlineWardleyMaps][] in a headless Chrome browser and captures the SVG and PNG of the created map.

```wardleymap
sequenceDiagram
    Actor A as User
    participant B as mermaid.go
    participant C as chromedp

    A ->>+ B: NewRenderEngine()
    B ->>+ C: Lanch new instance of chrome and eval JS library
    C -->> B: 
    B -->> A: 
    
    loop Render Process
        A ->> B: Render()
        B ->> C: mermaid.render()
        C ->> B: { svg, boxModel, exceptions }
        B ->> A: Result{ Svg, BoxModel Error }
    end

    A ->> B: Cancel()
    B -->> C: Context done
    C -->>- C: Shutdown chrome instance
    B -->>- A: 
```

Installation:

```shell
go install github.com/mrueg/go-wardley/cmd/owm
```

# CLI
An CLI is available [here](cmd/main.go).

# How to build

1. Checkout the code base
   `git clone https://github.com/mrueg/go-wardley.git`
2. Fetch the latest version of mermaid.js  
    `curl -LO https://unpkg.com/mermaid/dist/mermaid.min.js`
    Or if you want a specific version
    `curl -LO https://unpkg.com/mermaid@10.3.0/dist/mermaid.min.js`
3. Test it  
   `go test ./...`

# License

- [go-wardley][]: MIT License
- [onlinewardleymaps][]: MIT License
- [chromedp][]: MIT License
 
# Credits

This package is heavily inspired and reusing code from [mermaid.go](https://github.com/dreampuf/mermaid.go).

[go-wardley]: https://github.com/mrueg/go-wardley
[onlinewardleymaps]: https://github.com/damonsk/onlinewardleymaps
[chromedp]: https://github.com/chromedp/chromedp

