module github.com/blue-monads/turnix

go 1.24.0

require (
	github.com/alecthomas/kong v1.12.1
	github.com/alecthomas/repr v0.5.1
	github.com/cjoudrey/gluahttp v0.0.0-20201111170219-25003d9adfa9
	github.com/flosch/go-humanize v0.0.0-20140728123800-3ba51eabe506
	github.com/gin-gonic/gin v1.10.0
	github.com/gobwas/ws v1.4.0
	github.com/hako/branca v0.0.0-20200807062402-6052ac720505
	github.com/jaevor/go-nanoid v1.4.0
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/modelcontextprotocol/go-sdk v1.0.0
	github.com/ncruces/go-sqlite3 v0.30.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/psanford/sqlite3vfs v0.0.0-20240315230605-24e1d98cf361
	github.com/rqlite/sql v0.0.0-20250623131620-453fa49cad04
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/upper/db/v4 v4.7.0
	github.com/yuin/gopher-lua v1.1.1
	github.com/ztrue/tracerr v0.4.0
	golang.org/x/crypto v0.43.0
)

// replace github.com/psanford/sqlite3vfs => ../sqlite3vfs

// go get github.com/blue-monads/db/v4@mj-change-sqlite-driver

replace github.com/upper/db/v4 => github.com/blue-monads/db/v4 v4.0.0-20251111024918-ce036f164885

require (
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/eknkc/basex v1.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/segmentio/fasthash v1.0.3 // indirect
	github.com/tetratelabs/wazero v1.10.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
