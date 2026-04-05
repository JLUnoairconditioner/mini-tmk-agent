module mini-tmk-agent

go 1.21

require (
	github.com/joho/godotenv v1.5.1
	//github.com/pterm/pterm v0.12.66
	github.com/pterm/pterm v0.12.50
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/gookit/color v1.5.3 // indirect
	// github.com/inconshreveable/log15/v2 v2.3.2-0.20231026120738-e1f1e9b00bb8 // indirect
	// github.com/inconshreveable/log15/v3 v3.0.0-testing.5+incompatible // indirect
	// github.com/inconshreveable/loggers v0.0.0-20200515194955-c78405c5f5b3 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/term v0.12.0 // indirect
)

require github.com/gordonklaus/portaudio v0.0.0-20260203164431-765aa7dfa631

require (
	atomicgo.dev/cursor v0.1.1 // indirect
	atomicgo.dev/keyboard v0.2.8 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lithammer/fuzzysearch v1.1.5 // indirect
	golang.org/x/text v0.4.0 // indirect
)

//replace github.com/atomicgo/cursor => atomicgo.dev/cursor v0.2.0
//replace github.com/atomicgo/cursor => atomicgo.dev/cursor v0.2.0
replace github.com/atomicgo/cursor => atomicgo.dev/cursor v0.2.0

replace github.com/atomicgo/keyboard => atomicgo.dev/keyboard v0.2.9
