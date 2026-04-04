#! /bin/bash

echo -e "Start running the script..."
cd ../

echo -e "Start building the app for windows platform..."
NSIS_USER_BIN="$LOCALAPPDATA/NSIS3/Bin"
if [ -d "$NSIS_USER_BIN" ]; then
  export PATH="$NSIS_USER_BIN:$PATH"
fi
wails build --clean --platform windows/amd64 --nsis

echo -e "End running the script!"
