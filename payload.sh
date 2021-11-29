#!/bin/bash

mkdir -p /usr/local/etc

# use rm/install to create a new file with locked off permissions so a timing attack can't get a read handle
rm -f /usr/local/etc/ca.pem /usr/local/etc/ca_key.pem /usr/local/etc/localhost.pem /usr/local/etc/localhost_key.pem
install -m 644 /dev/null /usr/local/etc/ca.pem
install -m 600 /dev/null /usr/local/etc/ca_key.pem
install -m 644 /dev/null /usr/local/etc/localhost.pem
install -m 600 /dev/null /usr/local/etc/localhost_key.pem

# cat/heredoc doesn't leak process arguments
cat > /usr/local/etc/ca.pem << EOM
{{.CA -}}
EOM

cat > /usr/local/etc/ca_key.pem << EOM
{{.CAKey -}}
EOM

cat > /usr/local/etc/localhost.pem << EOM
{{.Localhost -}}
EOM

cat > /usr/local/etc/localhost_key.pem << EOM
{{.LocalhostKey -}}
EOM
