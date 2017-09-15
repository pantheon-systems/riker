#!/bin/bash
# shellcheck disable=SC2029

# Required environment vars:
#   OU: The certificate 'OU' field
#   CN: The certificate common-name (typically a DNS name or email addr)
#   FILENAME: Basename for cert and key files. eg: FILENAME=foo: foo.key, foo.crt, foo.pem

# Optional environment vars:
#   DIRECTORY: Directory to store downloaded certs in. Defaults to current dir ('.')
#   CA_HOST: Defaults to 'cimt.getpantheon.com' (production CA). Use your onebox address to create a development/sandbox cert.
#   SANS: SubjectAltNames to add to the cert. The CN will automatically be added to the SAN list so you don't need to add it.
#         The format is OpenSSL's SAN format documented here: https://www.openssl.org/docs/man1.0.2/apps/x509v3_config.html (Subject Alt Name section)
#         Example - 1 DNS name, 1 ipv4 address, 1 ipv6 address:
#           SANS="DNS:foo;IP:10.0.0.1:IP:2001::1"

# Usage examples:
#
#   - Create a certificate with CN=foo, OU=bar and no extra SANs (subjectAltNames) with filenames: mycert.key, mycert.crt, mycert.pem
#
#       CN=foo OU=bar FILENAME=mycert bash ./create-tls-cert.sh
#
#   - Add SubjectAltNames "foobar.com", and IP "10.0.0.1":
#
#       CN=foo OU=bar FILENAME=mycert SANS="DNS:foobar.com;IP:10.0.0.1" bash ./create-tls-cert.sh
#
#   - Issue a development certificate from a onebox (any onebox can be used, so use yours if you have one):
#
#       CA_HOST=onebox CN=foo OU=bar FILENAME=mycert bash ./create-tls-cert.sh

set -eou pipefail

CA_HOST="${CA_HOST:-cimt.getpantheon.com}"
DIRECTORY="${DIRECTORY:-.}"
FILENAME="${FILENAME:-}"
OU="${OU:-}"
CN="${CN:-}"
SANS="${SANS:-}"

if [[ -z "$CA_HOST" ]] || [[ -z "$FILENAME" ]] || [[ -z "$OU" ]] || [[ -z "$CN" ]]; then
    echo "missing one or more required env vars: CA_HOST, FILENAME, OU, CN"
    exit 1
fi

main() {
    echo "[INFO] SSH'ing to '$CA_HOST' to create MTLS certificate: CN=$CN, OU=$OU, FILENAME=$FILENAME, DIRECTORY=$DIRECTORY, SANS=$SANS"
    ssh "${CA_HOST}" "sudo pantheon pki.create_key:cn='$CN',ou='$OU',san=\"$SANS\",filename='$FILENAME',directory='.',noninteractive=True" >/dev/null

    echo "[INFO] Downloading $CA_HOST:$FILENAME.{key,crt,pem}"
    scp "${CA_HOST}:${FILENAME}.{key,crt,pem}" "$DIRECTORY/" >/dev/null
    ssh "${CA_HOST}" "sudo rm -f -- ${FILENAME}.{key,crt,pem}" >/dev/null

    echo "[INFO] Downloaded MTLS certificate files (Run 'openssl x509 -text -noout -in $FILENAME.pem' to view certificate):"
    ls -l "$DIRECTORY/$FILENAME".{key,crt,pem}
}
main "$@"
