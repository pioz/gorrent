case "$1" in
  "osx")
    mv ./deploy/darwin/gorrent.app ./deploy/darwin/Gorrent.app
    sips -s format icns ./qrc/donkey.png --out ./deploy/darwin/Gorrent.app/Contents/Resources/gorrent.icns
    mkdir ./deploy/darwin/Gorrent
    mv ./deploy/darwin/Gorrent.app ./deploy/darwin/Gorrent
    hdiutil create -volname Gorrent -srcfolder ./deploy/darwin/Gorrent -ov -format UDZO ./releases/Gorrent-osx-`cat VERSION`.dmg
    ;;
  "win32")
    echo "TODO"
    ;;
  "win64")
    echo "TODO"
    ;;
  *)
    echo "usage: $0 (osx|win32|win64)"
esac

