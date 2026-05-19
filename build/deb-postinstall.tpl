#!/bin/bash

if type update-alternatives 2>/dev/null >&1; then
    # Remove previous link if it doesn't use update-alternatives
    if [ -L '/usr/bin/doraterm' -a -e '/usr/bin/doraterm' -a "`readlink '/usr/bin/doraterm'`" != '/etc/alternatives/doraterm' ]; then
        rm -f '/usr/bin/doraterm'
    fi
    update-alternatives --install '/usr/bin/doraterm' 'doraterm' '/opt/Dora/doraterm' 100 || ln -sf '/opt/Dora/doraterm' '/usr/bin/doraterm'
else
    ln -sf '/opt/Dora/doraterm' '/usr/bin/doraterm'
fi

chmod 4755 '/opt/Dora/chrome-sandbox' || true

if hash update-mime-database 2>/dev/null; then
    update-mime-database /usr/share/mime || true
fi

if hash update-desktop-database 2>/dev/null; then
    update-desktop-database /usr/share/applications || true
fi
