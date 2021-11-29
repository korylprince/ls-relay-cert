package main

import "time"

type Config struct {
	MDMPrefix           string        `required:"true"`
	MDMToken            string        `required:"true"`
	SigningIdentity     string        `required:"true"`
	CacheSize           int           `default:"1024"`
	CacheTTL            time.Duration `default:"5m"`
	CachePrefix         string        `required:"true"`
	PayloadVersion      int           `default:"1"`
	PayloadIdentifier   string        `default:"com.github.korylprince.ls-relay-cert"`
	PayloadUUID         string        `required:"true"`
	PayloadOrganization string        `required:"true"`
	ListenAddr          string        `default:":80"`
}
