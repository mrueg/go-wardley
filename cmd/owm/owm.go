package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/mrueg/go-wardley/wardley"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd := &cli.Command{
		Name:                  "owm",
		Usage:                 "Render Wardley Maps on the command line",
		Version:               fmt.Sprintf("%s - %s@%s", version, commit, date),
		EnableShellCompletion: true,
		HideHelpCommand:       true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Value:       0,
				Aliases:     []string{"p"},
				Usage:       "Port the internal communication runs on",
				DefaultText: "random",
				Action: func(ctx context.Context, cmd *cli.Command, v int) error {
					if v < 0 || v >= 65536 {
						return fmt.Errorf("port value %v is out of range [0-65535]", v)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:      "input",
				Value:     "",
				Aliases:   []string{"i"},
				Usage:     "Path to file that contains the Wardley Map (if unset, read from stdin)",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:        "output",
				Value:       "",
				Aliases:     []string{"o"},
				DefaultText: "map.svg / map.png",
				Usage:       "Path of the output file",
			},
			&cli.StringFlag{
				Name:    "format",
				Value:   "svg",
				Aliases: []string{"f"},
				Usage:   "Output format. Possible values: svg, png",
			},
			&cli.FloatFlag{
				Name:    "scale",
				Value:   1.0,
				Aliases: []string{"s"},
				Usage:   "For PNG output only, scaling factor of the rendered image",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
				Usage: "Set log level: error, warn, info, debug, trace",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					level, err := zerolog.ParseLevel(s)
					if err != nil {
						return err
					}
					zerolog.SetGlobalLevel(level)
					return nil
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			err := Run(cmd)
			if err != nil {
				log.Error().Err(err).Msg("")
				os.Exit(1)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

func Run(cmd *cli.Command) error {
	port := strconv.FormatInt(int64(cmd.Int("port")), 10)
	re, _ := wardley.NewRenderEngine(port)
	defer re.Cancel()

	file := cmd.String("input")

	var content, outputContent []byte
	var err error
	if file == "" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		content, err = os.ReadFile(file)
		if err != nil {
			return err
		}
	}

	output := cmd.String("output")
	switch cmd.String("format") {

	case "svg":
		log.Info().Msg("Rendering SVG")
		outputContent, err = re.Render(string(content))
		if err != nil {
			return err
		}
		if output == "" {
			output = "map.svg"
		}
	case "png":
		log.Info().Float64("scale", cmd.Float64("scale")).Msg("Rendering PNG")
		outputContent, _, err = re.RenderAsScaledPng(string(content), cmd.Float64("scale"))

		if err != nil {
			return err
		}
		if output == "" {
			output = "map.png"
		}
	default:
		return fmt.Errorf("unknown output format")
	}

	err = os.WriteFile(output, outputContent, 0644)
	if err != nil {
		return err
	}
	log.Info().Str("file", output).Msg("Content written")
	return nil
}
