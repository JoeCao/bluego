module bluego

require (
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-gonic/gin v1.7.7
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/googollee/go-engine.io v1.4.1
	github.com/googollee/go-socket.io v1.4.2
	github.com/karalabe/hid v1.0.0
	github.com/muka/go-bluetooth v0.0.0-20190905083735-68fa9c3514a2
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/suapapa/go_eddystone v0.0.0-20190827074641-8d8c1bb79363 // indirect
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

go 1.13
