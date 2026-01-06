module github.com/blue-monads/turnix

go 1.24.1

toolchain go1.24.11

require (
	github.com/alecthomas/kong v1.12.1
	github.com/alecthomas/repr v0.5.1
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cjoudrey/gluahttp v0.0.0-20201111170219-25003d9adfa9
	github.com/flosch/go-humanize v0.0.0-20140728123800-3ba51eabe506
	github.com/gin-gonic/gin v1.10.0
	github.com/gobwas/ws v1.4.0
	github.com/hako/branca v0.0.0-20200807062402-6052ac720505
	github.com/jaevor/go-nanoid v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/modelcontextprotocol/go-sdk v1.0.0
	github.com/nbd-wtf/go-nostr v0.52.3
	github.com/ncruces/go-sqlite3 v0.30.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/psanford/sqlite3vfs v0.0.0-20240315230605-24e1d98cf361
	github.com/rqlite/sql v0.0.0-20251204023435-65660522892e
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/spf13/afero v1.15.0
	github.com/tidwall/gjson v1.18.0
	github.com/tidwall/pretty v1.2.1
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/upper/db/v4 v4.7.0
	github.com/yuin/gopher-lua v1.1.1
	github.com/ztrue/tracerr v0.4.0
	golang.org/x/crypto v0.43.0
	golang.org/x/net v0.45.0
	layeh.com/gopher-json v0.0.0-20201124131017-552bb3c4c3bf
)

// replace github.com/psanford/sqlite3vfs => ../sqlite3vfs

replace github.com/upper/db/v4 => github.com/blue-monads/db/v4 v4.0.0-20251111024918-ce036f164885

require (
	github.com/FactomProject/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/FactomProject/btcutilecc v0.0.0-20130527213604-d3a63a5752ec // indirect
	github.com/ImVexed/fasturl v0.0.0-20230304231329-4e41488060f3 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/bytedance/sonic v1.13.1 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/coder/websocket v1.8.12 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/eknkc/basex v1.0.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/segmentio/fasthash v1.0.3 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tetratelabs/wazero v1.10.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/tyler-smith/go-bip32 v1.0.0 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/arch v0.15.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/protobuf v1.36.2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
