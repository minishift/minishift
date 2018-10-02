#!/bin/sh

# Copyright (C) 2018 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ -z "$GOPATH" ]; then
    echo GOPATH environment variable not set
    exit
fi

ICON_DIR=$GOPATH/src/github.com/minishift/minishift/icons
SCRIPTS_DIR=$GOPATH/src/github.com/minishift/minishift/scripts
ICON_PKG_DIR=$GOPATH/src/github.com/minishift/minishift/pkg/minishift/systemtray/icon

TRAY_ICON_UNIX=$ICON_DIR/trayicon_unix.png
TRAY_ICON_WIN=$ICON_DIR/trayicon_win.ico
DOES_NOT_EXIST_ICON=$ICON_DIR/doesnotexist_unix.png
RUNNING_ICON=$ICON_DIR/running_unix.png
STOPPED_ICON=$ICON_DIR/stopped_unix.png

if [ ! -e "$GOPATH/bin/2goarray" ]; then
    echo "Installing 2goarray..."
    go get github.com/cratonica/2goarray
    if [ $? -ne 0 ]; then
        echo Failure executing go get github.com/cratonica/2goarray
        exit
    fi
fi

# $1 - go file name
# $2 - byte array name
# $3 - icon file name
function generate_go_package() {
	echo '// File generated using minishift/scripts/systemtray/gen_systemtray_icons.sh' >> $1
	echo '// Icon/png files are in minishift/icons directory' >> $1
	echo >> $1

	cat "$3" | $GOPATH/bin/2goarray $2 icon >> $1
	if [ $? -ne 0 ]; then
	    echo Failure generating $1
	    exit
	fi
	sed -e :a -e '/^\n*$/{$d;N;};/\n$/ba' -i $1 #remove trailing new line
	sed 's/[ \t]*$//' -i $1 #remove trailing white space
	echo Finished generating $1
}

function add_license() {
	sed 's/COPYRIGHTHOLDER/Copyright (C) 2018 Red Hat, Inc/' $SCRIPTS_DIR/boilerplate/boilerplate.go.txt >> $1
	if [ $? -ne 0 ]; then
	    echo Failure
	    exit
	fi
	echo Finished adding license
}

OUTPUT_TRAY_ICON_UNIX=$ICON_PKG_DIR/trayicon_unix.go
OUTPUT_TRAY_ICON_WIN=$ICON_PKG_DIR/trayicon_windows.go
# only unix icons atm, as we don't have status icons for windows yet
OUTPUT_DNE_ICON=$ICON_PKG_DIR/doesnotexist_unix.go
OUTPUT_RUNNING_ICON=$ICON_PKG_DIR/running_unix.go
OUTPUT_STOPPED_ICON=$ICON_PKG_DIR/stopped_unix.go

files=($OUTPUT_TRAY_ICON_UNIX $OUTPUT_TRAY_ICON_WIN $OUTPUT_DNE_ICON $OUTPUT_RUNNING_ICON $OUTPUT_STOPPED_ICON)
for f in ${files[@]}; do
	rm -rf $f
done

echo "// +build !windows" > $OUTPUT_TRAY_ICON_UNIX
echo >> $OUTPUT_TRAY_ICON_UNIX
add_license $OUTPUT_TRAY_ICON_UNIX
generate_go_package $OUTPUT_TRAY_ICON_UNIX "TrayIcon" $TRAY_ICON_UNIX

echo "// +build !darwin" > $OUTPUT_TRAY_ICON_WIN
echo >> $OUTPUT_TRAY_ICON_WIN
add_license $OUTPUT_TRAY_ICON_WIN
generate_go_package $OUTPUT_TRAY_ICON_WIN "TrayIcon" $TRAY_ICON_WIN

file_list=($OUTPUT_DNE_ICON $OUTPUT_RUNNING_ICON $OUTPUT_STOPPED_ICON)
for icon_file in ${file_list[@]}; do
	add_license $icon_file
done

generate_go_package $OUTPUT_DNE_ICON "DoesNotExist" $DOES_NOT_EXIST_ICON
generate_go_package $OUTPUT_RUNNING_ICON "Running" $RUNNING_ICON
generate_go_package $OUTPUT_STOPPED_ICON "Stopped" $STOPPED_ICON
