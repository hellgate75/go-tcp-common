#!/bin/bash
# call this script with an email address (valid or not).
# like:
# ./makecert.sh hellgate75@gmail.com

function usage() {
	echo -e "makecert.sh [options] | [-d | --default]"
	echo -e "  -d or --default      Accepts default parameters and continue"
	echo -e "Options: "
	echo -e "  -c or --country      Flag of country (default: \"$COUNTRY\")"
	echo -e "  -s or --state        Flag of your state (default: \"$STATE\")"
	echo -e "  -o or --organization Name of your organization (default: \"$ORGANIZATION\")"
	echo -e "  -u or --unit         Name of your organization Unit (default: \"$UNIT\")"
	echo -e "  -n or --cn           Name of your common name (default: \"$CN\")"
	echo -e "  -e or --email        Given email (default: \"$EMAIL\")"
}

COUNTRY="IE"
STATE="DUB"
ORGANIZATION="My Organization"
UNIT="IT"
#CN="$(hostname)"
CN="localhost"
EMAIL="hellgate75@gmail.com"
if [ "$#" -lt 1 ]; then
	echo "No parameter provided!!"
	usage
	exit
fi

if [ "-d" != "$1" ] && [ "--default" != "$1" ]; then
	while [ "$1" != "" ]; do
	    case $1 in
	        -c | --country )   		shift
	                                COUNTRY="$1"
	                                ;;
	        -s | --state )   		shift
	                                STATE="$1"
	                                ;;
	        -o | --organization )   shift
	                                ORGANIZATION="$1"
	                                ;;
	        -u | --unit )   		shift
	                                UNIT="$1"
	                                ;;
	        -n | --cn )   			shift
	                                CN="$1"
	                                ;;
	        -e | --email )   		shift
	                                EMAIL="$1"
	                                ;;
	        -h | --help )           usage
	                                exit
	                                ;;
	        * )                     usage
	                                exit 1
	    esac
	    shift
	done
fi

echo -e "Summary:\n  Country: $COUNTRY\n  State: $STATE\n  Unit: $UNIT\n  CN: $CN\n  E-Mail: $EMAIL"
if [ ! -e certs ]; then
	mkdir certs
fi 
rm certs/* 2> /dev/null
echo "create CA"
openssl genrsa -out certs/ca.key 2048
openssl req -new -x509 -key certs/ca.key -out certs/ca.crt -subj "$(echo "/C=$COUNTRY/ST=$STATE/L=Earth/O=\"$ORGANIZATION\"/OU=$UNIT/CN=$CN/emailAddress=$EMAIL")"
echo "make server cert"
openssl req -new -nodes -x509 -out certs/server.pem -keyout certs/server.key -days 3650 -subj "$(echo "/C=$COUNTRY/ST=$STATE/L=Earth/O=\"$ORGANIZATION\"/OU=$UNIT/CN=$CN/emailAddress=$EMAIL")"
echo "make client cert"
openssl req -new -nodes -x509 -in certs/ca.crt -inform PEM -out certs/client.pem -keyout certs/client.key -days 3650 -subj "$(echo "/C=$COUNTRY/ST=$STATE/L=Earth/O=\"$ORGANIZATION\"/OU=$UNIT/CN=$CN/emailAddress=$EMAIL")"

if [ "Linux" == "$(uname -s)" ]; then
	sudo cp certs/ca.crt /usr/share/ca-certificate/
	cd /usr/share/ca-certificate
	sudo chmod 644 certificate.crt
	sudo dpkg-reconfigure ca-certificates
	sudo update-ca-certificates
else
	echo "Please install CA certificate as trusted root certification authority, in order to prevent X509 truststore problems"
fi