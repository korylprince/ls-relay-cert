# About

gen-ls-cert is a utility for generating certificates for Lightspeed Relay Smart Agent. This utility is similar to Lightspeed Systems' `makeca` utility, but it's open source and it will also generate Configuration Profile (.mobileconfig) for trusting the CA certificate via an MDM.

# Installing

You can download A universal binary for macOS [here](https://github.com/korylprince/ls-relay-cert/releases/tag/v1.3.1). To build it yourself, you just need Go installed in $PATH:

```bash
GOBIN="$(pwd)" go install "github.com/korylprince/ls-relay-cert/cmd/gen-ls-cert@v1.3.1"
./gen-ls-cert -h
```

# Usage

```
Usage of gen-ls-cert:
  -identifier string
    	The top level profile identifier, and a prefix for the inner payload (default "com.github.korylprince.ls-relay-cert")
  -org string
    	The organization used for the profile (default "Lightspeed Systems")
  -out string
    	Output directory (default ".")
  -uuid string
    	The UUID used for the profile (default "randomly generated")
  -version int
    	The version used for the profile (default 1)
  -years int
    	The number of years to use for the CA (default 10)
```
