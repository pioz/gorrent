case "$1" in
  "osx")
    rm -r deploy/darwin
    qtdeploy build darwin
    mv deploy/darwin/gorrent.app deploy/darwin/Gorrent.app
    sips -s format icns qrc/donkey.png --out deploy/darwin/Gorrent.app/Contents/Resources/gorrent.icns
    mkdir deploy/darwin/Gorrent
    mv deploy/darwin/Gorrent.app deploy/darwin/Gorrent
    hdiutil create -volname Gorrent -srcfolder deploy/darwin/Gorrent -ov -format UDZO releases/Gorrent-osx-`cat VERSION`.dmg
    ;;
  "win32")
    ffmpeg -i qrc/donkey.png -vf scale=256:256 gorrent.ico
    echo "IDI_ICON1 ICON DISCARDABLE \"gorrent.ico\"" > icon.rc
    windres -i icon.rc -o icon_windows.syso -F pe-i386
    rm -r deploy/windows
    qtdeploy -docker build windows_32_static
    #makecert -sv pioz.pvk -n "CN=Enrico Pilotto,E=epilotto@gmx.com" -r pioz.cer
    #pvk2pfx -pvk pioz.pvk -spc pioz.cer -pfx pioz.pfx -po $PFX_PSW
    signtool sign -t http://timestamp.verisign.com/scripts/timstamp.dll -f ~/certs/pioz.pfx -p $PFX_PSW -debug deploy/windows/gorrent.exe
    zip -jr releases/Gorrent-win32-`cat VERSION`.zip deploy/windows/gorrent.exe
    rm icon.rc icon_windows.syso gorrent.ico
    ;;
  "win64")
    ffmpeg -i qrc/donkey.png -vf scale=256:256 gorrent.ico
    echo "IDI_ICON1 ICON DISCARDABLE \"gorrent.ico\"" > icon.rc
    windres -i icon.rc -o icon_windows.syso -F pe-x86-64
    rm -r deploy/windows
    qtdeploy -docker build windows_64_static
    signtool sign -t http://timestamp.verisign.com/scripts/timstamp.dll -f gorrent.pfx -p $PFX_PSW -debug deploy/windows/gorrent.exe
    zip -jr releases/Gorrent-win32-`cat VERSION`.zip deploy/windows/gorrent.exe
    rm icon.rc icon_windows.syso gorrent.ico
    ;;
  *)
    echo "usage: $0 (osx|win32|win64)"
esac
