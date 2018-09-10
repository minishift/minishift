// +build linux

package systray

/*
#cgo linux pkg-config: gtk+-3.0 appindicator3-0.1

#include "systray_linux.h"
*/
import "C"
import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
)

func nativeLoop() (err error) {
	_, err = C.nativeLoop()
	return
}

func quit() {
	C.quit()
}

// SetIcon sets the systray icon.
// iconBytes should be the content of .ico for windows and .ico/.jpg/.png
// for other platforms.
func setIcon(iconBytes []byte) error {
	bh := md5.Sum(iconBytes)
	dataHash := hex.EncodeToString(bh[:])
	iconFilePath := filepath.Join(os.TempDir(), "systray_temp_icon_"+dataHash)

	if _, err := os.Stat(iconFilePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(iconFilePath, iconBytes, 0644); err != nil {
			return err
		}
	}

	return setIconPath(iconFilePath)
}

// SetIcon sets the systray icon.
// iconBytes should be the content of .ico for windows and .ico/.jpg/.png
// for other platforms.
func setIconPath(path string) (err error) {
	_, err = C.setIcon(C.CString(path))
	return err
}

// SetTitle sets the systray title, only available on Mac.
func setTitle(title string) (err error) {
	_, err = C.setTitle(C.CString(title))
	return
}

// SetTooltip sets the systray tooltip to display on mouse hover of the tray icon,
// only available on Mac and Windows.
func setTooltip(tooltip string) (err error) {
	_, err = C.setTooltip(C.CString(tooltip))
	return
}

func addOrUpdateMenuItem(item *MenuItem) (err error) {
	if item.isSeparator {
		return addSeparator(item.id)
	}
	var disabled, checkable, checked C.short
	if item.disabled {
		disabled = 1
	}
	if item.checkable {
		checkable = 1
	}
	if item.checked {
		checked = 1
	}
	_, err = C.add_or_update_menu_item(
		C.int(item.id),
		C.CString(item.title),
		C.CString(item.tooltip),
		disabled,
		checkable,
		checked,
	)
	return
}

func addSeparator(id int32) (err error) {
	_, err = C.add_separator(C.int(id))
	return
}

func hideMenuItem(item *MenuItem) (err error) {
	_, err = C.hide_menu_item(
		C.int(item.id),
	)
	return
}

func showMenuItem(item *MenuItem) (err error) {
	_, err = C.show_menu_item(
		C.int(item.id),
	)
	return
}

//export systray_ready
func systray_ready() {
	go systrayReady()
}

//export systray_on_exit
func systray_on_exit() {
	systrayExit()
}

//export systray_menu_item_selected
func systray_menu_item_selected(cID C.int, cChecked C.int) {
	systrayMenuItemSelected(int32(cID), true, int32(cChecked) != 0)
}

// stubs for submenu and butmap menu items
func addBitmap(bmpbyte []byte, item *MenuItem) error {
	return nil
}

func addBitmapPath(fp string, item *MenuItem) error {
	return nil
}

func createSubMenu(id int32) {
}

func addSubmenuToTray(item *MenuItem) {
}
