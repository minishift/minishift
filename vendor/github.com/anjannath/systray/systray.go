/*
Package systray is a cross-platform Go library to place an icon and menu in the
notification area. Based on https://github.com/getlantern/systray and provides few improvements:

	* Windows system calls instead of custom DLL
	* Caching icon files
	* Allows to set direct path to icon file
	* Error handling

Supports Windows, Mac OSX and Linux currently.
Methods can be called from any goroutine except Run(), which should be called
at the very beginning of main() to lock at main thread.
*/
package systray

import (
	"errors"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

var (
	hasStarted = int32(0)
	hasQuit    = int32(0)
)

const (
	ItemDefault        = 0
	ItemSeparator byte = 1 << iota
	ItemChecked
	ItemCheckable
	ItemDisabled
)

// MenuItem is used to keep track each menu item of systray
// Don't create it directly, use the one systray.AddMenuItem() returned
type MenuItem struct {
	// ClickedCh is the channel which will be notified when the menu item is clicked
	clickedCh chan struct{}

	// id uniquely identify a menu item, not supposed to be modified
	id int32
	// menuId uniquely identify a menu the menu item belogns to
	menuId int32
	// title is the text shown on menu item
	title string
	// tooltip is the text shown when pointing to menu item
	tooltip string

	isSeparator   bool
	isSubmenuItem bool
	isSubmenu     bool

	checkable bool
	checked   bool
	disabled  bool
}

var (
	systrayReady  func()
	systrayExit   func()
	menuItems     = make(map[int32]*MenuItem)
	menuItemsLock sync.RWMutex

	iconTitle   string
	iconTooltip string
	iconPath    string
	iconData    []byte
	idSequence  = int32(-1)
)

// Run initializes GUI and starts the event loop, then invokes the onReady
// callback.
// It blocks until systray.Quit() is called.
// Should be called at the very beginning of main() to lock at main thread.
func Run(onReady func(), onExit func()) error {
	runtime.LockOSThread()

	systrayReady = func() {
		atomic.StoreInt32(&hasStarted, 1)
		initMenu()
		if onReady != nil {
			onReady()
		}
	}

	// onExit runs in the event loop to make sure it has time to
	// finish before the process terminates
	if onExit == nil {
		onExit = func() {}
	}
	systrayExit = onExit

	return nativeLoop()
}

// Quit the systray
func Quit() {
	if atomic.LoadInt32(&hasStarted) == 1 && atomic.CompareAndSwapInt32(&hasQuit, 0, 1) {
		quit()
	}
}

// AddMenuItem adds menu item with designated title and tooltip, returning a channel
// that notifies whenever that menu item is clicked.
//
// It can be safely invoked from different goroutines.
func AddMenuItem(title, tooltip string, flags byte) *MenuItem {
	menuItemsLock.Lock()
	defer menuItemsLock.Unlock()

	if ItemSeparator&flags != 0 {
		// remove other flags if separator
		flags = ItemSeparator
	} else if ItemCheckable&flags == 0 {
		// remove "checked" flag if not checkable
		flags &= ^ItemChecked
	}

	item := &MenuItem{
		clickedCh:     make(chan struct{}),
		id:            atomic.AddInt32(&idSequence, 1),
		title:         title,
		tooltip:       tooltip,
		isSubmenuItem: false,
		isSubmenu:     false,
		isSeparator:   ItemSeparator&flags != 0,
		checkable:     ItemCheckable&flags != 0,
		checked:       ItemChecked&flags != 0,
		disabled:      ItemDisabled&flags != 0,
	}
	menuItems[item.id] = item

	if atomic.LoadInt32(&hasStarted) == 1 {
		item.update()
	}
	return item
}

// AddSubMenu adds a sub menu to the systray and
// and returns an id to access to accesss the menu
func AddSubMenu(title string) *MenuItem {
	subMenuId := atomic.AddInt32(&idSequence, 1)
	createSubMenu(subMenuId)

	item := &MenuItem{
		clickedCh:     make(chan struct{}),
		id:            atomic.AddInt32(&idSequence, 1),
		title:         title,
		menuId:        subMenuId,
		isSubmenuItem: false,
		isSubmenu:     true,
	}
	if atomic.LoadInt32(&hasStarted) == 1 {
		addSubmenuToTray(item)
	}
	return item
}

func (sitem *MenuItem) AddSubMenuItem(title, tooltip string, flags byte) *MenuItem {
	menuItemsLock.Lock()
	defer menuItemsLock.Unlock()

	if ItemSeparator&flags != 0 {
		// remove other flags if separator
		flags = ItemSeparator
	} else if ItemCheckable&flags == 0 {
		// remove "checked" flag if not checkable
		flags &= ^ItemChecked
	}

	item := &MenuItem{
		clickedCh:     make(chan struct{}),
		id:            atomic.AddInt32(&idSequence, 1),
		title:         title,
		tooltip:       tooltip,
		menuId:        sitem.menuId,
		isSeparator:   ItemSeparator&flags != 0,
		isSubmenuItem: true,
		isSubmenu:     false,
		checkable:     ItemCheckable&flags != 0,
		checked:       ItemChecked&flags != 0,
		disabled:      ItemDisabled&flags != 0,
	}
	menuItems[item.id] = item

	if atomic.LoadInt32(&hasStarted) == 1 {
		item.update()
	}
	return item
}

func (item *MenuItem) AddBitmap(bitmapByte []byte) error {
	menuItemsLock.Lock()
	defer menuItemsLock.Unlock()

	if atomic.LoadInt32(&hasStarted) == 1 {
		return addBitmap(bitmapByte, item)
	}
	return errors.New("Could not set bitmap on menuitem")
}

func (item *MenuItem) AddBitmapPath(filepath string) error {
	menuItemsLock.Lock()
	defer menuItemsLock.Unlock()

	if atomic.LoadInt32(&hasStarted) == 1 {
		return addBitmapPath(filepath, item)
	}
	return errors.New("Could not set bitmap on menuitem")
}

// AddSeparator adds a separator bar to the menu.
func AddSeparator() *MenuItem {
	return AddMenuItem("", "", ItemSeparator)
}

// SetTitle set the text to display on a menu item.
func (item *MenuItem) SetTitle(title string) error {
	item.title = title
	return item.update()
}

// SetTooltip set the tooltip to show when mouse hover.
func (item *MenuItem) SetTooltip(tooltip string) error {
	item.tooltip = tooltip
	return item.update()
}

// Disabled checkes if the menu item is disabled.
func (item *MenuItem) Disabled() bool {
	return item.disabled
}

// Enable a menu item regardless if it's previously enabled or not.
func (item *MenuItem) Enable() error {
	item.disabled = false
	return item.update()
}

// Disable a menu item regardless if it's previously disabled or not.
func (item *MenuItem) Disable() error {
	item.disabled = true
	return item.update()
}

// Hide hides a menu item.
func (item *MenuItem) Hide() error {
	return hideMenuItem(item)
}

// Show shows a previously hidden menu item.
func (item *MenuItem) Show() error {
	return showMenuItem(item)
}

// Checked returns if the menu item has a check mark.
func (item *MenuItem) Checked() bool {
	return item.checked
}

// Check a menu item regardless if it's previously checked or not.
func (item *MenuItem) Check() error {
	if !item.checkable {
		return errors.New("item must be checkable")
	}
	item.checked = true
	return item.update()
}

// Uncheck a menu item regardless if it's previously unchecked or not.
func (item *MenuItem) Uncheck() error {
	if !item.checkable {
		return errors.New("item must be checkable")
	}
	item.checked = false
	return item.update()
}

// update propogates changes on a menu item to systray.
func (item *MenuItem) update() error {
	return addOrUpdateMenuItem(item)
}

func (item *MenuItem) OnClickCh() <-chan struct{} {
	return item.clickedCh
}

// SetIcon sets the systray icon.
// iconBytes should be the content of *.ico for windows and *.ico/*.jpg/*.png
// for other platforms.
// On windows and linux it creates reusable temporary file with icon content. This file
func SetIcon(iconBytes []byte) error {
	if atomic.LoadInt32(&hasStarted) == 0 {
		iconData = iconBytes
		return nil
	}
	return setIcon(iconBytes)
}

// SetIconPath sets icon by direct path.
// It should be *.ico for windows and *.ico/*.jpg/*.png for other platforms.
func SetIconPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	if atomic.LoadInt32(&hasStarted) == 0 {
		iconPath = path
		return nil
	}
	return setIconPath(path)
}

// SetTitle sets the systray title, only available on Mac.
func SetTitle(title string) error {
	if atomic.LoadInt32(&hasStarted) == 0 {
		iconTitle = title
		return nil
	}
	return setTitle(title)
}

// SetTooltip sets the systray tooltip to display on mouse hover of the tray icon,
// only available on Mac and Windows.
func SetTooltip(tooltip string) error {
	if atomic.LoadInt32(&hasStarted) == 0 {
		iconTooltip = tooltip
		return nil
	}
	return setTooltip(tooltip)
}

func initMenu() {
	menuItemsLock.RLock()
	defer menuItemsLock.RUnlock()

	if iconPath != "" {
		SetIconPath(iconPath)
	} else if iconData != nil {
		SetIcon(iconData)
	}

	if iconTitle != "" {
		SetTitle(iconTitle)
	}

	if iconTooltip != "" {
		SetTooltip(iconTooltip)
	}

	keys := make([]int, len(menuItems))
	i := 0
	for k := range menuItems {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)
	for _, k := range keys {
		menuItems[int32(k)].update()
	}
}

func systrayMenuItemSelected(id int32, forceState bool, newState bool) {
	menuItemsLock.RLock()
	defer menuItemsLock.RUnlock()
	item, ok := menuItems[id]
	if !ok {
		return
	}

	if item.checkable {
		if forceState {
			item.checked = newState
		} else {
			// toggle state
			item.checked = !item.checked
			item.update()
		}
	}

	select {
	case item.clickedCh <- struct{}{}:
		// in case no one waiting for the channel
	default:
	}
}
