#!/bin/bash
FOLDER="/home/ubuntu"

# Create a directory
mkdir -p $FOLDER/test1
mkdir -p $FOLDER/test
# Remove created directory
rmdir $FOLDER/test1

# FILE - A variable that holds the value of the file you want to test.
# You can reuse this same code block to test different files 
  # by changing the value of the $FILE variable.
FILE="test.txt"

# The -e operator tests for file existence
if [ -e $FILE ]
then
  # If the operator returns true, print a message saying the file exists.
  echo "$FILE exists"
else
  # If the operator returns true, print a message saying the file doesn't exist,
    # then creates the file with the name you defined in the FILE variable.
  echo "$FILE does not exist, creating new file" && touch test.txt
  
fi

# Create a file
touch test1.txt
# Write to the file
echo "test1 written line 1" > test1.txt
touch test2.txt
echo "test2 written line" > test2.txt

echo "test1 written line 2" >> test1.txt
echo "test2 re-written line 1" > test2.txt
