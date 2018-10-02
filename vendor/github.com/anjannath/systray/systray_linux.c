#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <limits.h>
#include <libappindicator/app-indicator.h>
#include "systray_linux.h"

static AppIndicator *global_app_indicator;
static GtkWidget *global_tray_menu = NULL;
static GList *global_menu_items = NULL;

typedef struct {
	GtkWidget *menu_item;
	int menu_id;
} MenuItemNode;

typedef struct {
	int menu_id;
	char* title;
	char* tooltip;
	short disabled;
	short checkable;
	short checked;
} MenuItemInfo;

int nativeLoop(void) {
	gtk_init(0, NULL);
	global_app_indicator = app_indicator_new("systray", "",
			APP_INDICATOR_CATEGORY_APPLICATION_STATUS);
	app_indicator_set_status(global_app_indicator, APP_INDICATOR_STATUS_ACTIVE);
	global_tray_menu = gtk_menu_new();
	app_indicator_set_menu(global_app_indicator, GTK_MENU(global_tray_menu));
	systray_ready();
	gtk_main();
	systray_on_exit();
	return 0;
}

void _systray_menu_item_selected(int *id) {
	systray_menu_item_selected(*id, 0);
}

void _systray_menu_item_changed(int *id) {
    int checked = 0;
    GList* it;
	for(it = global_menu_items; it != NULL; it = it->next) {
        MenuItemNode* item = (MenuItemNode*)(it->data);
        if(item->menu_id == *id){
            checked = gtk_check_menu_item_get_active((GtkCheckMenuItem*)(item->menu_item)) ? 1 : 0;
            break;
        }
    }
	systray_menu_item_selected(*id, checked);
}

// runs in main thread, should always return FALSE to prevent gtk to execute it again
gboolean do_add_or_update_menu_item(gpointer data) {
	MenuItemInfo *mii = (MenuItemInfo*)data;
	GList* it;
	for(it = global_menu_items; it != NULL; it = it->next) {
		MenuItemNode* item = (MenuItemNode*)(it->data);
		if(item->menu_id == mii->menu_id){
			gtk_menu_item_set_label(GTK_MENU_ITEM(item->menu_item), mii->title);
			if(mii->checkable == 1) {
			    gtk_check_menu_item_set_active((GtkCheckMenuItem*)(item->menu_item), mii->checked == 1);
			}
			break;
		}
	}

	// menu id doesn't exist, add new item
	if(it == NULL) {
		GtkWidget *menu_item;
		int *id = malloc(sizeof(int));
		*id = mii->menu_id;
		if(mii->checkable) {
		    menu_item = gtk_check_menu_item_new_with_label(mii->title);
		    gtk_check_menu_item_set_active((GtkCheckMenuItem*)(menu_item), mii->checked == 1);
		    g_signal_connect_swapped(G_OBJECT(menu_item), "toggled", G_CALLBACK(_systray_menu_item_changed), id);
		} else {
		    menu_item = gtk_menu_item_new_with_label(mii->title);
		    g_signal_connect_swapped(G_OBJECT(menu_item), "activate", G_CALLBACK(_systray_menu_item_selected), id);
		}
		gtk_menu_shell_append(GTK_MENU_SHELL(global_tray_menu), menu_item);

		MenuItemNode* new_item = malloc(sizeof(MenuItemNode));
		new_item->menu_id = mii->menu_id;
		new_item->menu_item = menu_item;
		GList* new_node = malloc(sizeof(GList));
		new_node->data = new_item;
		new_node->next = global_menu_items;
		if(global_menu_items != NULL) {
			global_menu_items->prev = new_node;
		}
		global_menu_items = new_node;
		it = new_node;
	}
	GtkWidget * menu_item = GTK_WIDGET(((MenuItemNode*)(it->data))->menu_item);
	gtk_widget_set_sensitive(menu_item, mii->disabled == 1 ? FALSE : TRUE);
	gtk_widget_show(menu_item);

	free(mii->title);
	free(mii->tooltip);
	free(mii);
	return FALSE;
}

gboolean do_add_separator(gpointer data) {
	GtkWidget *separator = gtk_separator_menu_item_new();
	gtk_menu_shell_append(GTK_MENU_SHELL(global_tray_menu), separator);
	gtk_widget_show(separator);
	return FALSE;
}

// runs in main thread, should always return FALSE to prevent gtk to execute it again
gboolean do_hide_menu_item(gpointer data) {
	MenuItemInfo *mii = (MenuItemInfo*)data;
	GList* it;
	for(it = global_menu_items; it != NULL; it = it->next) {
		MenuItemNode* item = (MenuItemNode*)(it->data);
		if(item->menu_id == mii->menu_id){
			gtk_widget_hide(GTK_WIDGET(item->menu_item));
			break;
		}
	}
	return FALSE;
}

// runs in main thread, should always return FALSE to prevent gtk to execute it again
gboolean do_show_menu_item(gpointer data) {
	MenuItemInfo *mii = (MenuItemInfo*)data;
	GList* it;
	for(it = global_menu_items; it != NULL; it = it->next) {
		MenuItemNode* item = (MenuItemNode*)(it->data);
		if(item->menu_id == mii->menu_id){
			gtk_widget_show(GTK_WIDGET(item->menu_item));
			break;
		}
	}
	return FALSE;
}

// runs in main thread, should always return FALSE to prevent gtk to execute it again
gboolean do_quit(gpointer data) {
	// app indicator doesn't provide a way to remove it, hide it as a workaround
	app_indicator_set_status(global_app_indicator, APP_INDICATOR_STATUS_PASSIVE);
	gtk_main_quit();
	return FALSE;
}

void setIcon(char* icon_path) {
	app_indicator_set_icon_full(global_app_indicator, icon_path, "");
	app_indicator_set_attention_icon_full(global_app_indicator, icon_path, "");
	free(icon_path);
}

void setTitle(char* ctitle) {
	app_indicator_set_label(global_app_indicator, ctitle, "");
	free(ctitle);
}

void setTooltip(char* ctooltip) {
	free(ctooltip);
}

void add_or_update_menu_item(int menu_id, char* title, char* tooltip, short disabled, short checkable, short checked) {
	MenuItemInfo *mii = malloc(sizeof(MenuItemInfo));
	mii->menu_id = menu_id;
	mii->title = title;
	mii->tooltip = tooltip;
	mii->disabled = disabled;
	mii->checkable = checkable;
	mii->checked = checked;
	g_idle_add(do_add_or_update_menu_item, mii);
}

void add_separator(int menu_id) {
	MenuItemInfo *mii = malloc(sizeof(MenuItemInfo));
	mii->menu_id = menu_id;
	g_idle_add(do_add_separator, mii);
}

void hide_menu_item(int menu_id) {
	MenuItemInfo *mii = malloc(sizeof(MenuItemInfo));
	mii->menu_id = menu_id;
	g_idle_add(do_hide_menu_item, mii);
}

void show_menu_item(int menu_id) {
	MenuItemInfo *mii = malloc(sizeof(MenuItemInfo));
	mii->menu_id = menu_id;
	g_idle_add(do_show_menu_item, mii);
}

void quit() {
	g_idle_add(do_quit, NULL);
}
