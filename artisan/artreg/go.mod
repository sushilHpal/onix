module github.com/gatblau/onix/artisan/artreg

go 1.16

replace (
	github.com/gatblau/onix/artisan => ../
	github.com/gatblau/onix/oxlib => ../../oxlib
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/gatblau/onix/artisan v0.0.0-00010101000000-000000000000
	github.com/gatblau/onix/oxlib v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-uuid v1.0.2
	github.com/minio/minio-go/v7 v7.0.21
	github.com/prometheus/client_golang v1.11.0
	github.com/swaggo/http-swagger v1.1.1
	github.com/swaggo/swag v1.7.1
	gopkg.in/yaml.v2 v2.4.0
)
