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

mod=(`echo $opArg | sed 's/@/\n/g'`)

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
echo -e "\n$BYellow[Log]$BGreen Working directory: $DIR$Black"

micros=(actions admin auth notifications profile setting storage)

# get latest tag
get_latest_tag() {
  
    local outpu=$(curl -s GET "https://api.github.com/repos/$1/tags?per_page=1" | # Get latest tag from GitHub api
    grep '"name":' |                                                                              # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/')                                                                # Pluck JSON value
    echo $outpu
}

# Update telar-core module
update_telar_core(){
ver="${1}"
if [ -z "$ver" ];
then
    ver=$(get_latest_tag "red-gold/telar-core")
fi
echo -e "$BYellow[Log]$BGreen Updating telar-core version $ver $Black"
for i in "${micros[@]}"; 
do 
microPath="$DIR/micros/$i"
echo -e "$BYellow[Log]$BCyan Updating [telar-core] module in $microPath"
cd "$microPath" && go get github.com/red-gold/telar-core@$ver && go mod tidy
echo -e "$BYellow[Log]$BCyan [telar-core] module updated for $i$Black"
done
}


# Update telar-web module
update_telar_web(){
ver="${1}"
if [ -z "$ver" ];
then
    ver=$(get_latest_tag "red-gold/telar-web")
fi
echo -e "$BYellow[Log]$BGreen Updating telar-web version $ver $Black"
for i in "${micros[@]}"; 
do 
microPath="$DIR/micros/$i"
echo -e "$BYellow[Log]$BCyan Updating [telar-web] module in $microPath"
cd "$microPath" && go get github.com/red-gold/telar-web@$ver && go mod tidy
echo -e "$BYellow[Log]$BCyan Updating [telar-web] module updated for $i$Black"
done
}

case ${mod[0]} in

  telar-core)
    echo -e "$BYellow[Log]$BCyan Updating [telar-core] module ..."
    update_telar_core ${mod[1]}
    ;;
  telar-web)
    echo -e "$BYellow[Log]$BCyan Updating [telar-web] module"
    update_telar_web ${mod[1]}
    ;;

  *)
    echo -n "unknown"
    ;;
esac