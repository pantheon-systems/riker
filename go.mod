module github.com/pantheon-systems/riker

require (
	github.com/Sirupsen/logrus v1.0.4
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v0.0.0-20171215095238-967bee733a73
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jroimartin/gocui v0.0.0-20170827195011-4f518eddb04b
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/nlopes/slack v0.0.0-20181230171726-db1f92067f7b
	github.com/nsf/termbox-go v0.0.0-20190325093121-288510b9734e // indirect
	github.com/pantheon-systems/go-certauth v0.0.0-20170606170341-8764720d23a5
	github.com/pantheon-systems/riker/health v0.0.0
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/cobra v0.0.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.0
	github.com/spf13/viper v1.0.0
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734 // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/sys v0.0.0-20190502175342-a43fa875dd82 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20190502173448-54afdca5d873 // indirect
	google.golang.org/grpc v1.19.0
)

replace github.com/pantheon-systems/riker/health => ./health
