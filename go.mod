module github.com/EduOJ/judgeServer

go 1.14

require (
	github.com/EduOJ/backend v0.0.0-20210302111520-6e9f42f0290c
	github.com/EduOJ/judger v0.0.5
	github.com/go-redis/redis/v8 v8.6.0 // indirect
	github.com/go-resty/resty/v2 v2.4.0
	github.com/klauspost/cpuid/v2 v2.0.3 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.10 // indirect
	github.com/minio/sha256-simd v0.1.1
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/stretchr/testify v1.6.1 => github.com/leoleoasd/testify v1.6.2-0.20200818074144-885db91dbfe9
