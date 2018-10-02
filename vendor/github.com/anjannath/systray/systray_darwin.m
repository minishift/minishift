#import <Cocoa/Cocoa.h>
#include "systray_darwin.h"

@interface MenuItem : NSObject
{
  @public
    NSNumber* menuId;
    NSNumber* submenuId;
    NSString* title;
    NSString* tooltip;
    short disabled;
    short checked;
    short isSubmenu;
    short isSubmenuItem;
}
-(id) initWithId: (int)theMenuId
   withSubmenuId: (int)theSubmenuId
       withTitle: (const char*)theTitle
     withTooltip: (const char*)theTooltip
    withDisabled: (short)theDisabled
     withChecked: (short)theChecked
   withIsSubmenu: (short)theisSubmenu
withIsSubmenuItem: (short)theisSubmenuItem;
     @end
     @implementation MenuItem
     -(id) initWithId: (int)theMenuId
	withSubmenuId: (int)theSubmenuId
            withTitle: (const char*)theTitle
          withTooltip: (const char*)theTooltip
         withDisabled: (short)theDisabled
          withChecked: (short)theChecked
        withIsSubmenu: (short)theisSubmenu
    withIsSubmenuItem: (short)theisSubmenuItem
{
  menuId = [NSNumber numberWithInt:theMenuId];
  submenuId = [NSNumber numberWithInt:theSubmenuId];
  title = [[NSString alloc] initWithCString:theTitle
                                   encoding:NSUTF8StringEncoding];
  tooltip = [[NSString alloc] initWithCString:theTooltip
                                     encoding:NSUTF8StringEncoding];
  disabled = theDisabled;
  checked = theChecked;
  isSubmenu = theisSubmenu;
  isSubmenuItem = theisSubmenuItem;
  return self;
}
@end

@interface AppDelegate: NSObject <NSApplicationDelegate>
  - (void) add_or_update_menu_item:(MenuItem*) item;
  - (IBAction)menuHandler:(id)sender;
  @property (assign) IBOutlet NSWindow *window;
  @end

  @implementation AppDelegate
{
  NSStatusItem *statusItem;
  NSMenu *menu;
  NSMutableDictionary *submenus;
  NSImage* imageBuffer;
  NSCondition* cond;
}

@synthesize window = _window;

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification
{
  self->statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
  self->menu = [[NSMenu alloc] init];
  self->submenus = [@{} mutableCopy];
  [self->menu setAutoenablesItems: FALSE];
  [self->statusItem setMenu:self->menu];
  systray_ready();
}

- (void)applicationWillTerminate:(NSNotification *)aNotification
{
  systray_on_exit();
}

- (void)setIcon:(NSImage *)image {
  [statusItem setImage:image];
}

- (void)setTitle:(NSString *)title {
  [statusItem setTitle:title];
}

- (void)setTooltip:(NSString *)tooltip {
  [statusItem setToolTip:tooltip];
}

- (IBAction)menuHandler:(id)sender
{
  NSNumber* menuId = [sender representedObject];
  systray_menu_item_selected(menuId.intValue);
}

- (void) add_or_update_menu_item:(MenuItem*) item
{
  NSMenuItem* menuItem;
  NSMenu* parentMenu = menu;
  if (item->isSubmenuItem == 1){
	  parentMenu = submenus[item->submenuId];
    NSLog(@"*** submenuID: %@ ***", item->submenuId);
  }
  int existedMenuIndex = [parentMenu indexOfItemWithRepresentedObject: item->menuId];
  if (existedMenuIndex == -1){
    menuItem = [parentMenu addItemWithTitle:item->title action:@selector(menuHandler:) keyEquivalent:@""];
    NSLog(@"\nNew item: %@, menuItem: %@", item, menuItem);
    [menuItem setTarget:self];
    [menuItem setRepresentedObject: item->menuId];

  } else {
    menuItem = [parentMenu itemAtIndex: existedMenuIndex];
    NSLog(@"\nOld item: %@, menuItem: %@", item, menuItem);
    [menuItem setTitle:item->title];
  }
  [menuItem setToolTip:item->tooltip];
  if (item->disabled == 1) {
    NSLog(@"\nDisabling");
    [menuItem setEnabled:FALSE];
  } else {
    NSLog(@"\nEnabling");
    [menuItem setEnabled:TRUE];
  }
  if (item->checked == 1) {
    [menuItem setState:NSOnState];
  } else {
    [menuItem setState:NSOffState];
  }
}

- (void) add_sub_menu:(MenuItem*) item 
{
  NSMenuItem* menuItem;
  NSMenu* subMenu;
  int existedMenuIndex = [menu indexOfItemWithRepresentedObject: item->menuId];
  if (existedMenuIndex == -1) {
    menuItem = [menu addItemWithTitle:item->title action:@selector(menuHandler:) keyEquivalent:@""];
    subMenu = [[NSMenu alloc] init];
    [menuItem setTarget:self];
    [menuItem setSubmenu: subMenu];
    [menuItem setRepresentedObject: item->menuId];
    [submenus setObject:subMenu forKey:item->submenuId];

  } else {
    menuItem = [menu itemAtIndex: existedMenuIndex];
    [menuItem setTitle:item->title];
  }
  [menuItem setToolTip:item->tooltip];
  if (item->disabled == 1) {
    [menuItem setEnabled:FALSE];
  } else {
    [menuItem setEnabled:TRUE];
  }
  if (item->checked == 1) {
    [menuItem setState:NSOnState];
  } else {
    [menuItem setState:NSOffState];
  }
}

- (void) add_separator:(NSNumber*) menuId
{
  NSMenu* parentMenu = menu;
  if (submenus[menuId] != nil) {
    parentMenu = submenus[menuId];
  }
  [parentMenu addItem: [NSMenuItem separatorItem]];
}

- (void) hide_menu_item:(NSNumber*) menuId
{
  NSMenuItem* menuItem;
  NSMenu* parentMenu = menu;
  int existedMenuIndex = [parentMenu indexOfItemWithRepresentedObject: menuId];
  if (existedMenuIndex == -1) {
    return;
  }
  menuItem = [parentMenu itemAtIndex: existedMenuIndex];
  [menuItem setHidden:TRUE];
}

- (void) show_menu_item:(NSNumber*) menuId
{
  NSMenuItem* menuItem;
  NSMenu* parentMenu;
  if (submenus[menuId] != nil) {
    parentMenu = submenus[menuId];
  }
  int existedMenuIndex = [parentMenu indexOfItemWithRepresentedObject: menuId];
  if (existedMenuIndex == -1) {
    return;
  }
  menuItem = [parentMenu itemAtIndex: existedMenuIndex];
  [menuItem setHidden:FALSE];
}

- (void) load_bitmap_to_image_buffer:(NSImage *) image 
{
	imageBuffer = image;
}

- (void) add_bitmap_to_menu_item:(NSNumber*) menuId
{
  NSMenuItem* menuItem;
  NSMenu* parentMenu = menu;
  if (submenus[menuId] != nil) {
	  parentMenu = submenus[menuId];
  }
  int existedMenuIndex = [parentMenu indexOfItemWithRepresentedObject: menuId];
  if (existedMenuIndex == -1) {
    return;
  }
  menuItem = [parentMenu itemAtIndex: existedMenuIndex];
  [menuItem setImage:imageBuffer];
  
}

- (void) quit
{
  [NSApp terminate:self];
}

@end

int nativeLoop(void) {
  AppDelegate *delegate = [[AppDelegate alloc] init];
  [[NSApplication sharedApplication] setDelegate:delegate];
  [NSApp run];
  return EXIT_SUCCESS;
}

void runInMainThread(SEL method, id object) {
  [(AppDelegate*)[NSApp delegate]
    performSelectorOnMainThread:method
                     withObject:object
                  waitUntilDone: YES];
}

void setIcon(const char* iconBytes, int length) {
  NSData* buffer = [NSData dataWithBytes: iconBytes length:length];
  NSImage *image = [[NSImage alloc] initWithData:buffer];
  [image setSize:NSMakeSize(16, 16)];
  runInMainThread(@selector(setIcon:), (id)image);
}

void setTitle(char* ctitle) {
  NSString* title = [[NSString alloc] initWithCString:ctitle
                                             encoding:NSUTF8StringEncoding];
  free(ctitle);
  runInMainThread(@selector(setTitle:), (id)title);
}

void setTooltip(char* ctooltip) {
  NSString* tooltip = [[NSString alloc] initWithCString:ctooltip
                                               encoding:NSUTF8StringEncoding];
  free(ctooltip);
  runInMainThread(@selector(setTooltip:), (id)tooltip);
}

void add_or_update_menu_item(int menuId, int submenuId, char* title, char* tooltip, short disabled, short checked, short isSubmenu, short isSubmenuItem) {
  MenuItem* item = [[MenuItem alloc] initWithId: menuId 
				  withSubmenuId: submenuId
				      withTitle: title 
				    withTooltip: tooltip 
				   withDisabled: disabled 
				    withChecked: checked
				  withIsSubmenu: isSubmenu
			      withIsSubmenuItem: isSubmenuItem];
  free(title);
  free(tooltip);
  runInMainThread(@selector(add_or_update_menu_item:), (id)item);
}

void add_sub_menu(int menuId, int submenuId, char* title, char* tooltip, short disabled, short checked, short isSubmenu, short isSubmenuItem) {
  MenuItem* item = [[MenuItem alloc] initWithId: menuId 
				  withSubmenuId: submenuId
				      withTitle: title 
				    withTooltip: tooltip 
				   withDisabled: disabled 
				    withChecked: checked
				  withIsSubmenu: isSubmenu
			      withIsSubmenuItem: isSubmenuItem];
  free(title);
  free(tooltip);
  runInMainThread(@selector(add_sub_menu:), (id)item);
}

void add_separator(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(add_separator:), (id)mId);
}

void hide_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(hide_menu_item:), (id)mId);
}

void show_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(show_menu_item:), (id)mId);
}

void loadBitmapToImageBuffer(const char* bitmapBytes, int length) {
  NSData* buffer = [NSData dataWithBytes: bitmapBytes length:length];
  NSImage *image = [[NSImage alloc] initWithData:buffer];
  [image setSize:NSMakeSize(16, 16)];
  runInMainThread(@selector(load_bitmap_to_image_buffer:), (id)image);
}

void add_bitmap_to_menu_item(const char* bitmapBytes, int lenght, int menuId) {
  loadBitmapToImageBuffer(bitmapBytes, lenght);
  NSLog(@"Done loading image");
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  runInMainThread(@selector(add_bitmap_to_menu_item:), (id)mId);
  NSLog(@"Done setting image");
}


void quit() {
  runInMainThread(@selector(quit), nil);
}
