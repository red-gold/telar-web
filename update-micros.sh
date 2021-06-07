#!/bin/bash

# Colors                 
Black='\033[0;30m'        # Black
BRed='\033[1;31m'         # Red
BGreen='\033[1;32m'       # Green
BYellow='\033[1;33m'      # Yellow
BBlue='\033[1;34m'        # Blue
BPurple='\033[1;35m'      # Purple
BCyan='\033[1;36m'        # Cyan

opArg="${1}" 
arg="${2}"

echo Argument $arg1
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo -e "$BGreen Working directory: $DIR\n"
echo -e "$Black"

micros=(actions admin auth notifications profile setting storage)

# Update telar-core module
update_telar_core(){
for i in "${micros[@]}"; 
do 
echo -e "$BYellow"
microPath="$DIR/micros/$i"
echo "Updating [telar-core] module in $microPath"
cd "$microPath" && go get github.com/red-gold/telar-core@v0.1.12 && go mod tidy
echo -e "$BGreen"
echo "[telar-core] module updated for $i"
done
echo -e "$Black"
}


# Update telar-web module
update_telar_web(){
for i in "${micros[@]}"; 
do 
echo -e "$BYellow"
microPath="$DIR/micros/$i"
echo "Updating [telar-web] module in $microPath"
cd "$microPath" && go get github.com/red-gold/telar-web@v0.1.64 && go mod tidy
echo "[telar-web] module updated for $i"
done
echo -e "$Black"
}

case $opArg in

  telar-core)
    echo -n "Updating [telar-core] module ..."
    update_telar_core
    ;;
  telar-web)
    echo -n "Updating [telar-web] module"
    update_telar_web
    ;;

  *)
    echo -n "unknown"
    ;;
esac