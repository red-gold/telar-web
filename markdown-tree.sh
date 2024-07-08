# This function uses tree command to read all directories in the project root and save it in tree.md file.
# it logs the tree log before writing it to the file. 
read_all_directories(){
    echo -e "$BYellow[Log]$BCyan Start reading all directories in the project root"
    tree -d -L 4 > tree.md
    echo -e "$BYellow[Log]$BCyan Wrote all directories in the project root to tree.md file"
}

# run read_all_directories function
read_all_directories

# End of the script
