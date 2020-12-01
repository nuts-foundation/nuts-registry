module github.com/nuts-foundation/nuts-registry

go 1.14

require (
	github.com/cyberphone/json-canonicalization v0.0.0-20200417180520-cd6247b5f11e
	github.com/deepmap/oapi-codegen v1.4.1
	github.com/fsnotify/fsnotify v1.4.7
	github.com/golang/mock v1.4.4
	github.com/google/uuid v1.1.2
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/leodido/go-urn v1.2.1-0.20201207081027-996485e2f5f1
	github.com/lestrrat-go/jwx v1.0.5
	github.com/magiconair/properties v1.8.4
	github.com/nuts-foundation/nuts-crypto v0.15.1-0.20201113103650-0107d387c2e2
	github.com/nuts-foundation/nuts-go-core v0.15.0
	github.com/nuts-foundation/nuts-go-test v0.15.0
	github.com/nuts-foundation/nuts-network v0.15.2-0.20201113105043-f920bbeac58f
	github.com/pelletier/go-toml v1.5.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
)

replace github.com/leodido/go-urn => github.com/nuts-foundation/go-urn v1.2.1-0.20201207081027-996485e2f5f1
