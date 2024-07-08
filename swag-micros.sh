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

# function to run swag command(swag init -g router/route.go --parseDependency) for each micro direcory
run_swag_command(){
    for i in "${micros[@]}"; 
    do 
        microPath="$DIR/micros/$i"
        echo -e "$BYellow[Log]$BCyan Running swag command in $microPath"
        cd "$microPath" && swag init -g router/route.go --parseDependency
        echo -e "$BYellow[Log]$BCyan Swag command run for $i$Black"
    done
}

# run swag command(swag init -g router/route.go --parseDependency) for each micro direcory
run_swag_command

# A function that removes swag directory(keeps all swagger files) from root of project and create new swag directory. Then copy swagger.json and swagger.yaml.
# Then rename rename swagger.json to {micro name}.json and swagger.yaml to {micro name}.yaml
run_swag_directory(){
    echo -e "$BYellow[Log]$BCyan Removing swag directory from root of project"
    rm -rf "$DIR/swag"
    echo -e "$BYellow[Log]$BCyan Creating new swag directory"
    mkdir "$DIR/swag"
    for i in "${micros[@]}"; 
    do 
        echo -e "$BYellow[Log]$BCyan Copying swagger.json and swagger.yaml for $i"
        cp "$DIR/micros/$i/docs/swagger.json" "$DIR/swag/$i.json"
        cp "$DIR/micros/$i/docs/swagger.yaml" "$DIR/swag/$i.yaml"
        echo -e "$BYellow[Log]$BCyan Copied swagger.json and swagger.yaml for $i"
    done
}

# run swag_directory function
run_swag_directory

# a function that creates a json file name swagger.json. It will contain a list which contains 
# {api name: microservice name}, {api path: basePath} and {desctiption: info.description} for each micro.
create_swagger_json(){
    echo -e "$BYellow[Log]$BCyan Creating swagger.json"
    echo "[" > "$DIR/swag/swagger.json"
    first=1
    for i in "${micros[@]}"; 
    do 
        if [ "$first" -eq 1 ]; then
            first=0
        else
            echo -e "," >> "$DIR/swag/swagger.json"
        fi
        echo -e "$BYellow[Log]$BCyan Adding $i to swagger.json"
        echo -e "{\n\t\"name\": \"$i\",\n\t\"apiPath\": \"/$i\",\n\t\"description\": $(cat $DIR/micros/$i/docs/swagger.json | jq .info.description)\n}" >> "$DIR/swag/swagger.json"
        echo -e "$BYellow[Log]$BCyan Added $i to swagger.json"
    done
    echo "]" >> "$DIR/swag/swagger.json"
    echo -e "$BYellow[Log]$BCyan Created swagger.json"
}

# run create_swagger_json function
create_swagger_json


# a fuction zip swag directory and change the zip file to telar-web-swagger.zip
zip_swag_directory(){
    echo -e "$BYellow[Log]$BCyan Zipping swag directory"
    zip -r "$DIR/telar-web-swagger.zip" "$DIR/swag"
    echo -e "$BYellow[Log]$BCyan Zipped swag directory"
}

# run zip_swag_directory function
zip_swag_directory

# End of the script
# Example: ./swag-micros.sh